// Package mapper provides declarative flag-to-argument mapping for kubelogin converter
// This replaces the manual if-else chains in getArgValues() with a data-driven approach

package mapper

import (
	"github.com/Azure/kubelogin/pkg/internal/token"
)

// FlagMapping represents a declarative mapping between CLI flags and exec arguments
type FlagMapping struct {
	// FlagName is the CLI flag name (without dashes, e.g., "client-id")
	FlagName string
	
	// ArgumentName is the exec argument name (with dashes, e.g., "--client-id")
	ArgumentName string
	
	// GetValue extracts the value from TokenOptions
	GetValue func(*token.Options) string
	
	// LegacyConfigKey is the key in legacy authProvider config (optional)
	LegacyConfigKey string
	
	// IsRequired indicates if this flag is required for validation
	IsRequired bool
	
	// ApplicableLogins lists which login methods this flag applies to
	// Empty slice means applies to all login methods
	ApplicableLogins []string
	
	// IsBoolean indicates this is a boolean flag (no value, just presence)
	IsBoolean bool
	
	// GetBoolValue extracts boolean value from TokenOptions (for boolean flags)
	GetBoolValue func(*token.Options) bool
}

// Registry holds all flag mappings and provides lookup functionality
type Registry struct {
	mappings []FlagMapping
}

// NewRegistry creates a new flag mapping registry with default mappings
func NewRegistry() *Registry {
	return &Registry{
		mappings: defaultMappings(),
	}
}

// defaultMappings returns the complete set of flag-to-argument mappings
func defaultMappings() []FlagMapping {
	return []FlagMapping{
		{
			FlagName:     "client-id",
			ArgumentName: "--client-id",
			GetValue:     func(opts *token.Options) string { return opts.ClientID },
			LegacyConfigKey: "client-id",
			IsRequired:   true,
			ApplicableLogins: []string{
				token.DeviceCodeLogin,
				token.InteractiveLogin,
				token.ServicePrincipalLogin,
				token.ROPCLogin,
			},
		},
		{
			FlagName:     "server-id",
			ArgumentName: "--server-id",
			GetValue:     func(opts *token.Options) string { return opts.ServerID },
			LegacyConfigKey: "apiserver-id",
			IsRequired:   true,
			ApplicableLogins: []string{}, // Required for all login methods
		},
		{
			FlagName:     "tenant-id",
			ArgumentName: "--tenant-id",
			GetValue:     func(opts *token.Options) string { return opts.TenantID },
			LegacyConfigKey: "tenant-id",
			IsRequired:   true,
			ApplicableLogins: []string{
				token.DeviceCodeLogin,
				token.InteractiveLogin,
				token.ServicePrincipalLogin,
				token.ROPCLogin,
			},
		},
		{
			FlagName:     "environment",
			ArgumentName: "--environment",
			GetValue:     func(opts *token.Options) string { return opts.Environment },
			LegacyConfigKey: "environment",
			IsRequired:   false,
			ApplicableLogins: []string{
				token.DeviceCodeLogin,
				token.InteractiveLogin,
				token.ServicePrincipalLogin,
				token.ROPCLogin,
			},
		},
		{
			FlagName:     "login-hint",
			ArgumentName: "--login-hint",
			GetValue:     func(opts *token.Options) string { return opts.LoginHint },
			LegacyConfigKey: "", // No legacy equivalent
			IsRequired:   false,
			ApplicableLogins: []string{token.InteractiveLogin},
		},
		{
			FlagName:     "redirect-url",
			ArgumentName: "--redirect-url",
			GetValue:     func(opts *token.Options) string { return opts.RedirectURL },
			LegacyConfigKey: "", // No legacy equivalent
			IsRequired:   false,
			ApplicableLogins: []string{token.InteractiveLogin},
		},
		{
			FlagName:     "client-secret",
			ArgumentName: "--client-secret",
			GetValue:     func(opts *token.Options) string { return opts.ClientSecret },
			LegacyConfigKey: "",
			IsRequired:   false,
			ApplicableLogins: []string{token.ServicePrincipalLogin},
		},
		{
			FlagName:     "client-certificate",
			ArgumentName: "--client-certificate",
			GetValue:     func(opts *token.Options) string { return opts.ClientCert },
			LegacyConfigKey: "",
			IsRequired:   false,
			ApplicableLogins: []string{token.ServicePrincipalLogin},
		},
		{
			FlagName:     "client-certificate-password",
			ArgumentName: "--client-certificate-password",
			GetValue:     func(opts *token.Options) string { return opts.ClientCertPassword },
			LegacyConfigKey: "",
			IsRequired:   false,
			ApplicableLogins: []string{token.ServicePrincipalLogin},
		},
		{
			FlagName:     "username",
			ArgumentName: "--username",
			GetValue:     func(opts *token.Options) string { return opts.Username },
			LegacyConfigKey: "",
			IsRequired:   false,
			ApplicableLogins: []string{token.ROPCLogin},
		},
		{
			FlagName:     "password",
			ArgumentName: "--password",
			GetValue:     func(opts *token.Options) string { return opts.Password },
			LegacyConfigKey: "",
			IsRequired:   false,
			ApplicableLogins: []string{token.ROPCLogin},
		},
		{
			FlagName:     "identity-resource-id",
			ArgumentName: "--identity-resource-id",
			GetValue:     func(opts *token.Options) string { return opts.IdentityResourceID },
			LegacyConfigKey: "",
			IsRequired:   false,
			ApplicableLogins: []string{token.MSILogin},
		},
		{
			FlagName:     "authority-host",
			ArgumentName: "--authority-host",
			GetValue:     func(opts *token.Options) string { return opts.AuthorityHost },
			LegacyConfigKey: "",
			IsRequired:   false,
			ApplicableLogins: []string{token.WorkloadIdentityLogin},
		},
		{
			FlagName:     "federated-token-file",
			ArgumentName: "--federated-token-file",
			GetValue:     func(opts *token.Options) string { return opts.FederatedTokenFile },
			LegacyConfigKey: "",
			IsRequired:   false,
			ApplicableLogins: []string{token.WorkloadIdentityLogin},
		},
		{
			FlagName:     "cache-dir",
			ArgumentName: "--cache-dir",
			GetValue:     func(opts *token.Options) string { return opts.AuthRecordCacheDir },
			LegacyConfigKey: "",
			IsRequired:   false,
			ApplicableLogins: []string{}, // Applies to all
		},
		{
			FlagName:     "pop-claims",
			ArgumentName: "--pop-claims",
			GetValue:     func(opts *token.Options) string { return opts.PoPTokenClaims },
			LegacyConfigKey: "",
			IsRequired:   false,
			ApplicableLogins: []string{
				token.InteractiveLogin,
				token.ServicePrincipalLogin,
			},
		},
		// Boolean flags
		{
			FlagName:     "legacy",
			ArgumentName: "--legacy",
			IsBoolean:    true,
			GetBoolValue: func(opts *token.Options) bool { return opts.IsLegacy },
			LegacyConfigKey: "config-mode",
			IsRequired:   false,
			ApplicableLogins: []string{
				token.DeviceCodeLogin,
				token.ServicePrincipalLogin,
				token.ROPCLogin,
			},
		},
		{
			FlagName:     "pop-enabled",
			ArgumentName: "--pop-enabled",
			IsBoolean:    true,
			GetBoolValue: func(opts *token.Options) bool { return opts.IsPoPTokenEnabled },
			LegacyConfigKey: "",
			IsRequired:   false,
			ApplicableLogins: []string{
				token.InteractiveLogin,
				token.ServicePrincipalLogin,
			},
		},
		{
			FlagName:     "disable-environment-override",
			ArgumentName: "--disable-environment-override",
			IsBoolean:    true,
			GetBoolValue: func(opts *token.Options) bool { return opts.DisableEnvironmentOverride },
			LegacyConfigKey: "",
			IsRequired:   false,
			ApplicableLogins: []string{token.ServicePrincipalLogin},
		},
	}
}

// GetMappingsForLogin returns mappings applicable to the given login method
func (r *Registry) GetMappingsForLogin(loginMethod string) []FlagMapping {
	var result []FlagMapping
	for _, mapping := range r.mappings {
		if len(mapping.ApplicableLogins) == 0 || contains(mapping.ApplicableLogins, loginMethod) {
			result = append(result, mapping)
		}
	}
	return result
}

// GetMapping returns the mapping for a specific flag name
func (r *Registry) GetMapping(flagName string) (FlagMapping, bool) {
	for _, mapping := range r.mappings {
		if mapping.FlagName == flagName {
			return mapping, true
		}
	}
	return FlagMapping{}, false
}

// GetRequiredMappings returns mappings that are required for the given login method
func (r *Registry) GetRequiredMappings(loginMethod string) []FlagMapping {
	var result []FlagMapping
	for _, mapping := range r.GetMappingsForLogin(loginMethod) {
		if mapping.IsRequired {
			result = append(result, mapping)
		}
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
