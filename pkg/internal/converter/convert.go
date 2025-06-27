package converter

import (
	"fmt"
	"strings"

	"github.com/Azure/kubelogin/pkg/internal/converter/builder"
	"github.com/Azure/kubelogin/pkg/internal/converter/handlers"
	"github.com/Azure/kubelogin/pkg/internal/converter/mapper"
	"github.com/Azure/kubelogin/pkg/internal/token"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	klog "k8s.io/klog/v2"
)

const (
	azureAuthProvider = "azure"
	cfgClientID       = "client-id"
	cfgApiserverID    = "apiserver-id"
	cfgTenantID       = "tenant-id"
	cfgEnvironment    = "environment"
	cfgConfigMode     = "config-mode"

	argClientID                   = "--client-id"
	argServerID                   = "--server-id"
	argTenantID                   = "--tenant-id"
	argEnvironment                = "--environment"
	argClientSecret               = "--client-secret"
	argClientCert                 = "--client-certificate"
	argClientCertPassword         = "--client-certificate-password"
	argIsLegacy                   = "--legacy"
	argUsername                   = "--username"
	argPassword                   = "--password"
	argLoginMethod                = "--login"
	argIdentityResourceID         = "--identity-resource-id"
	argAuthorityHost              = "--authority-host"
	argFederatedTokenFile         = "--federated-token-file"
	argTokenCacheDir              = "--token-cache-dir"
	argAuthRecordCacheDir         = "--cache-dir"
	argIsPoPTokenEnabled          = "--pop-enabled"
	argPoPTokenClaims             = "--pop-claims"
	argDisableEnvironmentOverride = "--disable-environment-override"
	argRedirectURL                = "--redirect-url"
	argLoginHint                  = "--login-hint"

	flagAzureConfigDir             = "azure-config-dir"
	flagClientID                   = "client-id"
	flagContext                    = "context"
	flagServerID                   = "server-id"
	flagTenantID                   = "tenant-id"
	flagEnvironment                = "environment"
	flagClientSecret               = "client-secret"
	flagClientCert                 = "client-certificate"
	flagClientCertPassword         = "client-certificate-password"
	flagIsLegacy                   = "legacy"
	flagUsername                   = "username"
	flagPassword                   = "password"
	flagLoginMethod                = "login"
	flagIdentityResourceID         = "identity-resource-id"
	flagAuthorityHost              = "authority-host"
	flagFederatedTokenFile         = "federated-token-file"
	flagTokenCacheDir              = "token-cache-dir"
	flagAuthRecordCacheDir         = "cache-dir"
	flagIsPoPTokenEnabled          = "pop-enabled"
	flagPoPTokenClaims             = "pop-claims"
	flagDisableEnvironmentOverride = "disable-environment-override"
	flagRedirectURL                = "redirect-url"
	flagLoginHint                  = "login-hint"

	execName        = "kubelogin"
	getTokenCommand = "get-token"
	execAPIVersion  = "client.authentication.k8s.io/v1beta1"
	execInstallHint = `
kubelogin is not installed which is required to connect to AAD enabled cluster.

To learn more, please go to https://azure.github.io/kubelogin/
`

	azureConfigDir = "AZURE_CONFIG_DIR"
)

// buildConversionContext creates a ConversionContext from options and authInfo using the new mapping system
func buildConversionContext(o Options, authInfo *api.AuthInfo, registry *mapper.Registry) *handlers.ConversionContext {
	// Extract values from authInfo and options using the registry
	tokenOptions := &token.Options{
		LoginMethod: o.TokenOptions.LoginMethod,
	}

	// Populate token options using the flag registry mapping
	for _, mapping := range registry.GetAllMappings() {
		if mapping.IsBoolean {
			if o.isSet(mapping.FlagName) {
				// Use the flag value if explicitly set
				value := mapping.GetBoolValue(&o.TokenOptions)
				setBooleanField(tokenOptions, mapping.FlagName, value)
			} else {
				// Check for legacy auth provider config or existing exec args
				value := getBooleanValueFromAuthInfo(authInfo, mapping)
				setBooleanField(tokenOptions, mapping.FlagName, value)
			}
		} else {
			if o.isSet(mapping.FlagName) {
				// Use the flag value if explicitly set
				value := mapping.GetValue(&o.TokenOptions)
				if value != "" { // Only set non-empty values to avoid overwriting good values
					setStringField(tokenOptions, mapping.FlagName, value)
				} else if flagValue, err := o.Flags.GetString(mapping.FlagName); err == nil && flagValue != "" {
					// Use flag value directly if TokenOptions value is empty
					setStringField(tokenOptions, mapping.FlagName, flagValue)
				}
			} else {
				// Check for legacy auth provider config or existing exec args
				value := getStringValueFromAuthInfo(authInfo, mapping)
				setStringField(tokenOptions, mapping.FlagName, value)
			}
		}
	}

	return &handlers.ConversionContext{
		Options:          tokenOptions,
		AuthInfo:         authInfo,
		IsLegacyProvider: isLegacyAzureAuth(authInfo),
		FlagRegistry:     registry,
		IsSet:            o.isSet,
		AzureConfigDir:   o.azureConfigDir,
	}
}

// Helper functions to extract values from authInfo using mapping
func getStringValueFromAuthInfo(authInfo *api.AuthInfo, mapping mapper.FlagMapping) string {
	if authInfo == nil {
		return ""
	}

	isLegacyAuthProvider := isLegacyAzureAuth(authInfo)

	if isLegacyAuthProvider {
		if mapping.LegacyConfigKey != "" {
			if x, ok := authInfo.AuthProvider.Config[mapping.LegacyConfigKey]; ok {
				return x
			}
		}
	} else {
		result := getExecArg(authInfo, mapping.ArgumentName)
		// Special handling for cache-dir: also check for deprecated --token-cache-dir
		if result == "" && (mapping.FlagName == "cache-dir" || mapping.FlagName == "token-cache-dir") {
			result = getExecArg(authInfo, "--token-cache-dir")
		}
		return result
	}

	return ""
}

func getBooleanValueFromAuthInfo(authInfo *api.AuthInfo, mapping mapper.FlagMapping) bool {
	if authInfo == nil {
		return false
	}

	isLegacyAuthProvider := isLegacyAzureAuth(authInfo)

	if isLegacyAuthProvider {
		if mapping.LegacyConfigKey != "" {
			if x := authInfo.AuthProvider.Config[mapping.LegacyConfigKey]; x == "" || x == "0" {
				return true // Legacy mode logic
			}
		}
	} else {
		return getExecBoolArg(authInfo, mapping.ArgumentName)
	}

	return false
}

// Helper functions to set values on token options
func setStringField(options *token.Options, flagName, value string) {
	// For cache-dir and token-cache-dir, don't overwrite existing non-empty values
	if (flagName == "cache-dir" || flagName == "token-cache-dir") && options.AuthRecordCacheDir != "" && value == "" {
		return
	}

	switch flagName {
	case "client-id":
		options.ClientID = value
	case "server-id":
		options.ServerID = value
	case "tenant-id":
		options.TenantID = value
	case "environment":
		options.Environment = value
	case "client-secret":
		options.ClientSecret = value
	case "client-certificate":
		options.ClientCert = value
	case "client-certificate-password":
		options.ClientCertPassword = value
	case "username":
		options.Username = value
	case "password":
		options.Password = value
	case "identity-resource-id":
		options.IdentityResourceID = value
	case "authority-host":
		options.AuthorityHost = value
	case "federated-token-file":
		options.FederatedTokenFile = value
	case "cache-dir":
		options.AuthRecordCacheDir = value
	case "token-cache-dir": // Deprecated, but still supported
		options.AuthRecordCacheDir = value
	case "pop-claims":
		options.PoPTokenClaims = value
	case "redirect-url":
		options.RedirectURL = value
	case "login-hint":
		options.LoginHint = value
	}
}

func setBooleanField(options *token.Options, flagName string, value bool) {
	switch flagName {
	case "legacy":
		options.IsLegacy = value
	case "pop-enabled":
		options.IsPoPTokenEnabled = value
	case "disable-environment-override":
		options.DisableEnvironmentOverride = value
	}
}

func Convert(o Options, pathOptions *clientcmd.PathOptions) error {
	clientConfig := o.configFlags.ToRawKubeConfigLoader()
	var kubeconfigs []string

	klog.V(5).Info(o.ToString())

	if clientConfig.ConfigAccess() != nil {
		if clientConfig.ConfigAccess().GetExplicitFile() != "" {
			kubeconfigs = append(kubeconfigs, clientConfig.ConfigAccess().GetExplicitFile())
		} else {
			kubeconfigs = append(kubeconfigs, clientConfig.ConfigAccess().GetLoadingPrecedence()...)
		}
	}

	klog.V(5).Infof("Loading kubeconfig from %s", strings.Join(kubeconfigs, ":"))

	config, err := clientConfig.RawConfig()
	if err != nil {
		return fmt.Errorf("unable to load kubeconfig: %s", err)
	}

	targetAuthInfo := ""

	if o.context != "" {
		if config.Contexts[o.context] == nil {
			return fmt.Errorf("no context exists with the name: %q", o.context)
		}
		targetAuthInfo = config.Contexts[o.context].AuthInfo
	}

	for name, authInfo := range config.AuthInfos {

		if targetAuthInfo != "" && name != targetAuthInfo {
			continue
		}

		klog.V(5).Infof("context: %q", name)

		//  is it legacy aad auth or is it exec using kubelogin?
		if !isExecUsingkubelogin(authInfo) && !isLegacyAzureAuth(authInfo) {
			continue
		}

		klog.V(5).Info("converting...")

		// Create registry for flag mapping
		registry := mapper.NewRegistry()

		// Build conversion context using the new mapping system
		ctx := buildConversionContext(o, authInfo, registry)

		exec := &api.ExecConfig{
			Command: execName,
			Args: []string{
				getTokenCommand,
			},
			APIVersion:  execAPIVersion,
			InstallHint: execInstallHint,
		}

		// Preserve any existing install hint
		if authInfo.Exec != nil && authInfo.Exec.InstallHint != "" {
			exec.InstallHint = authInfo.Exec.InstallHint
		}

		// Don't add --login here, we'll add it at the end after all other arguments

		// Validate that server-id is available (required for all login methods)
		if ctx.Options.ServerID == "" {
			return fmt.Errorf("%s is required", argServerID)
		}

		// Cache directory and other arguments will be handled by the specific login method handlers

		// Use the new handler system to build login-method-specific arguments
		handlerRegistry := handlers.NewHandlerRegistry()
		handler, exists := handlerRegistry.GetHandler(o.TokenOptions.LoginMethod)
		if !exists {
			return fmt.Errorf("unsupported login method: %s", o.TokenOptions.LoginMethod)
		}

		// Validate the context for this login method
		validationResult := handler.Validate(ctx)
		if !validationResult.IsValid {
			return fmt.Errorf("%s", strings.Join(validationResult.Errors, ", "))
		}

		// Build arguments using the handler
		argBuilder := builder.NewExecArgsBuilder()
		err = handler.BuildExecArgs(ctx, argBuilder)
		if err != nil {
			return fmt.Errorf("failed to build exec args: %w", err)
		}

		// Get the built arguments and append them to exec.Args
		handlerArgs, err := argBuilder.Build()
		if err != nil {
			return fmt.Errorf("failed to build exec args: %w", err)
		}

		// Append all handler-built arguments
		exec.Args = append(exec.Args, handlerArgs...)

		// Add --login at the end (as expected by tests)
		exec.Args = append(exec.Args, argLoginMethod, o.TokenOptions.LoginMethod)

		// Handle special case for Azure CLI which needs environment variable
		if o.TokenOptions.LoginMethod == token.AzureCLILogin && o.azureConfigDir != "" {
			exec.Env = append(exec.Env, api.ExecEnvVar{Name: azureConfigDir, Value: o.azureConfigDir})
		}

		authInfo.Exec = exec
		authInfo.AuthProvider = nil
	}
	err = clientcmd.ModifyConfig(pathOptions, config, true)
	return err
}

// get the item in Exec.Args[] right after someArg
func getExecArg(authInfoPtr *api.AuthInfo, someArg string) (resultStr string) {
	if someArg == "" {
		return
	}
	if authInfoPtr == nil || authInfoPtr.Exec == nil || authInfoPtr.Exec.Args == nil {
		return
	}
	if len(authInfoPtr.Exec.Args) < 1 {
		return
	}
	for i := range authInfoPtr.Exec.Args {
		if authInfoPtr.Exec.Args[i] == someArg {
			if len(authInfoPtr.Exec.Args) > i+1 {
				return authInfoPtr.Exec.Args[i+1]
			}
		}
	}
	return
}

func getExecBoolArg(authInfoPtr *api.AuthInfo, someArg string) bool {
	if someArg == "" {
		return false
	}
	if authInfoPtr == nil || authInfoPtr.Exec == nil || authInfoPtr.Exec.Args == nil {
		return false
	}
	if len(authInfoPtr.Exec.Args) < 1 {
		return false
	}
	for i := range authInfoPtr.Exec.Args {
		if authInfoPtr.Exec.Args[i] == someArg {
			return true
		}
	}
	return false
}

// If enabling PoP token support, users must provide both "--pop-enabled" and "--pop-claims" flags together.
// If either is provided without the other, validation should throw an error, otherwise the get-token command
// will fail under the hood.
func validatePoPClaims(args []string, isPopTokenEnabled bool, popTokenClaimsFlag, popTokenClaimsVal string) ([]string, error) {
	if isPopTokenEnabled && popTokenClaimsVal == "" {
		// pop-enabled and pop-claims must be provided together
		return args, fmt.Errorf("%s is required when specifying %s", argPoPTokenClaims, argIsPoPTokenEnabled)
	}

	if popTokenClaimsVal != "" && !isPopTokenEnabled {
		// pop-enabled and pop-claims must be provided together
		return args, fmt.Errorf("%s is required when specifying %s", argIsPoPTokenEnabled, argPoPTokenClaims)
	}

	if isPopTokenEnabled && popTokenClaimsVal != "" {
		args = append(args, argIsPoPTokenEnabled)
		args = append(args, popTokenClaimsFlag, popTokenClaimsVal)
	}

	return args, nil
}

func isLegacyAzureAuth(authInfoPtr *api.AuthInfo) (ok bool) {
	if authInfoPtr == nil {
		return
	}
	if authInfoPtr.AuthProvider == nil {
		return
	}
	return authInfoPtr.AuthProvider.Name == azureAuthProvider
}

func isExecUsingkubelogin(authInfoPtr *api.AuthInfo) (ok bool) {
	if authInfoPtr == nil {
		return
	}
	if authInfoPtr.Exec == nil {
		return
	}
	lowerc := strings.ToLower(authInfoPtr.Exec.Command)
	return strings.Contains(lowerc, "kubelogin")
}
