// Package builder provides type-safe exec argument construction for kubelogin converter
// This replaces manual string slice construction with a fluent interface

package builder

import (
	"fmt"
	"strings"
)

// ExecArgsBuilder provides a fluent interface for building exec arguments
type ExecArgsBuilder struct {
	args   []string
	errors []error
}

// NewExecArgsBuilder creates a new exec args builder with the base command
func NewExecArgsBuilder() *ExecArgsBuilder {
	return &ExecArgsBuilder{
		args:   []string{"get-token"},
		errors: nil,
	}
}

// AddArgument adds a flag-value pair to the arguments
func (b *ExecArgsBuilder) AddArgument(flag, value string) *ExecArgsBuilder {
	if value == "" {
		return b // Skip empty values
	}
	b.args = append(b.args, flag, value)
	return b
}

// AddRequiredArgument adds a required flag-value pair, recording an error if value is empty
func (b *ExecArgsBuilder) AddRequiredArgument(flag, value string) *ExecArgsBuilder {
	if value == "" {
		b.errors = append(b.errors, fmt.Errorf("%s is required", flag))
		return b
	}
	b.args = append(b.args, flag, value)
	return b
}

// AddFlag adds a boolean flag (flag without value) if condition is true
func (b *ExecArgsBuilder) AddFlag(flag string, condition bool) *ExecArgsBuilder {
	if condition {
		b.args = append(b.args, flag)
	}
	return b
}

// AddOptionalArgument adds a flag-value pair only if value is not empty
func (b *ExecArgsBuilder) AddOptionalArgument(flag, value string) *ExecArgsBuilder {
	if value != "" {
		b.args = append(b.args, flag, value)
	}
	return b
}

// Build returns the final arguments slice and any accumulated errors
func (b *ExecArgsBuilder) Build() ([]string, error) {
	if len(b.errors) > 0 {
		return nil, b.combineErrors()
	}

	// Return a copy to prevent external modification
	result := make([]string, len(b.args))
	copy(result, b.args)
	return result, nil
}

// MustBuild returns the final arguments slice, panicking if there are errors
// This maintains compatibility with the original Build() method
func (b *ExecArgsBuilder) MustBuild() []string {
	args, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("ExecArgsBuilder validation failed: %v", err))
	}
	return args
}

// Length returns the current number of arguments
func (b *ExecArgsBuilder) Length() int {
	return len(b.args)
}

// Reset clears all arguments except the base command and clears all errors
func (b *ExecArgsBuilder) Reset() *ExecArgsBuilder {
	b.args = []string{"get-token"}
	b.errors = nil
	return b
}

// Clone creates a copy of the current builder including errors
func (b *ExecArgsBuilder) Clone() *ExecArgsBuilder {
	newBuilder := &ExecArgsBuilder{
		args:   make([]string, len(b.args)),
		errors: make([]error, len(b.errors)),
	}
	copy(newBuilder.args, b.args)
	copy(newBuilder.errors, b.errors)
	return newBuilder
}

// String returns a string representation of the current arguments
func (b *ExecArgsBuilder) String() string {
	if len(b.errors) > 0 {
		return fmt.Sprintf("ExecArgs[%d]: %v (errors: %d)", len(b.args), b.args, len(b.errors))
	}
	return fmt.Sprintf("ExecArgs[%d]: %v", len(b.args), b.args)
}

// HasErrors returns true if there are any validation errors
func (b *ExecArgsBuilder) HasErrors() bool {
	return len(b.errors) > 0
}

// GetErrors returns all accumulated errors
func (b *ExecArgsBuilder) GetErrors() []error {
	return b.errors
}

// combineErrors combines multiple errors into a single error message
func (b *ExecArgsBuilder) combineErrors() error {
	if len(b.errors) == 0 {
		return nil
	}

	if len(b.errors) == 1 {
		return b.errors[0]
	}

	var messages []string
	for _, err := range b.errors {
		messages = append(messages, err.Error())
	}

	return fmt.Errorf("multiple validation errors: %s", strings.Join(messages, "; "))
}

// Convenience builders for specific argument patterns

// RequiredAuthArgs is a convenience builder for common authentication arguments
type RequiredAuthArgs struct {
	ServerID string
	ClientID string
	TenantID string
}

// AddRequiredAuthArgs adds the standard required authentication arguments
func (b *ExecArgsBuilder) AddRequiredAuthArgs(args RequiredAuthArgs) *ExecArgsBuilder {
	return b.AddRequiredArgument("--server-id", args.ServerID).
		AddRequiredArgument("--client-id", args.ClientID).
		AddRequiredArgument("--tenant-id", args.TenantID)
}

// OptionalAuthArgs is a convenience builder for common optional authentication arguments
type OptionalAuthArgs struct {
	Environment string
	CacheDir    string
	IsLegacy    bool
}

// AddOptionalAuthArgs adds common optional authentication arguments
func (b *ExecArgsBuilder) AddOptionalAuthArgs(args OptionalAuthArgs) *ExecArgsBuilder {
	return b.AddOptionalArgument("--environment", args.Environment).
		AddOptionalArgument("--cache-dir", args.CacheDir).
		AddFlag("--legacy", args.IsLegacy)
}

// PoPTokenArgs is a convenience builder for PoP token arguments
type PoPTokenArgs struct {
	Enabled bool
	Claims  string
}

// AddPoPTokenArgs adds PoP token arguments with validation
// Both pop-enabled and pop-claims must be provided together
func (b *ExecArgsBuilder) AddPoPTokenArgs(args PoPTokenArgs) *ExecArgsBuilder {
	if args.Enabled && args.Claims == "" {
		b.errors = append(b.errors, fmt.Errorf("--pop-claims is required when --pop-enabled is specified"))
		return b
	}

	if !args.Enabled && args.Claims != "" {
		b.errors = append(b.errors, fmt.Errorf("--pop-enabled is required when --pop-claims is specified"))
		return b
	}

	if args.Enabled {
		b.AddFlag("--pop-enabled", true)
		b.AddRequiredArgument("--pop-claims", args.Claims)
	}

	return b
}

// InteractiveArgs is a convenience builder for interactive login specific arguments
type InteractiveArgs struct {
	RedirectURL string
	LoginHint   string
}

// AddInteractiveArgs adds interactive login specific arguments
func (b *ExecArgsBuilder) AddInteractiveArgs(args InteractiveArgs) *ExecArgsBuilder {
	return b.AddOptionalArgument("--redirect-url", args.RedirectURL).
		AddOptionalArgument("--login-hint", args.LoginHint)
}

// ServicePrincipalArgs is a convenience builder for service principal arguments
type ServicePrincipalArgs struct {
	ClientSecret               string
	ClientCertificate          string
	ClientCertificatePassword  string
	DisableEnvironmentOverride bool
}

// AddServicePrincipalArgs adds service principal specific arguments
func (b *ExecArgsBuilder) AddServicePrincipalArgs(args ServicePrincipalArgs) *ExecArgsBuilder {
	return b.AddOptionalArgument("--client-secret", args.ClientSecret).
		AddOptionalArgument("--client-certificate", args.ClientCertificate).
		AddOptionalArgument("--client-certificate-password", args.ClientCertificatePassword).
		AddFlag("--disable-environment-override", args.DisableEnvironmentOverride)
}

// ROPCArgs is a convenience builder for ROPC login arguments
type ROPCArgs struct {
	Username string
	Password string
}

// AddROPCArgs adds ROPC (Resource Owner Password Credentials) arguments
func (b *ExecArgsBuilder) AddROPCArgs(args ROPCArgs) *ExecArgsBuilder {
	return b.AddOptionalArgument("--username", args.Username).
		AddOptionalArgument("--password", args.Password)
}

// WorkloadIdentityArgs is a convenience builder for workload identity arguments
type WorkloadIdentityArgs struct {
	AuthorityHost      string
	FederatedTokenFile string
}

// AddWorkloadIdentityArgs adds workload identity specific arguments
func (b *ExecArgsBuilder) AddWorkloadIdentityArgs(args WorkloadIdentityArgs) *ExecArgsBuilder {
	return b.AddOptionalArgument("--authority-host", args.AuthorityHost).
		AddOptionalArgument("--federated-token-file", args.FederatedTokenFile)
}

// MSIArgs is a convenience builder for MSI (Managed Service Identity) arguments
type MSIArgs struct {
	IdentityResourceID string
}

// AddMSIArgs adds MSI specific arguments
func (b *ExecArgsBuilder) AddMSIArgs(args MSIArgs) *ExecArgsBuilder {
	return b.AddOptionalArgument("--identity-resource-id", args.IdentityResourceID)
}
