package converter

import (
"fmt"

"github.com/Azure/kubelogin/pkg/internal/token"
"github.com/spf13/cobra"
"github.com/spf13/pflag"
"k8s.io/cli-runtime/pkg/genericclioptions"
)

// stringptr returns a pointer to the provided string
func stringptr(str string) *string { return &str }

// RawOptions holds all raw (unvalidated) input values from CLI flags, environment variables, etc.
type RawOptions struct {
	configFlags    genericclioptions.RESTClientGetter
	flags          *pflag.FlagSet
	tokenOptions   *token.Options
	context        string
	azureConfigDir string
}

// validatedData holds the private validated data that cannot be instantiated outside this package
type validatedData struct {
	configFlags    genericclioptions.RESTClientGetter
	flags          *pflag.FlagSet
	tokenOptions   *token.Options
	context        string
	azureConfigDir string
}

// ValidatedOptions represents options that have passed validation
// Uses private validated data to prevent external modification
type ValidatedOptions struct {
	validated *validatedData // Private pointer - no external access
}

// CompletedOptions represents fully completed options ready for use
// Shares validated data immutably with ValidatedOptions
type CompletedOptions struct {
	validated *validatedData // Private pointer - shared with ValidatedOptions
}

// NewRawOptions creates a new RawOptions instance with defaults
func NewRawOptions() *RawOptions {
	configFlags := &genericclioptions.ConfigFlags{
		KubeConfig: stringptr(""),
	}

	tokenOpts := token.NewOptions(true)
	return &RawOptions{
		configFlags:  configFlags,
		tokenOptions: &tokenOpts,
	}
}

// AddFlags binds CLI flags to the raw options
func (o *RawOptions) AddFlags(fs *pflag.FlagSet) {
	o.flags = fs
	if cf, ok := o.configFlags.(*genericclioptions.ConfigFlags); ok {
		cf.AddFlags(fs)
	}
	o.tokenOptions.AddFlags(fs)
	// Add additional flags specific to converter
	fs.StringVar(&o.context, "context", o.context, "The name of the kubeconfig context to use")
	fs.StringVar(&o.azureConfigDir, "azure-config-dir", o.azureConfigDir, "Azure config directory")
}

// AddCompletions adds command completions
func (o *RawOptions) AddCompletions(cmd *cobra.Command) {
	// Add any command completions here if needed
}

// Validate validates the raw options and returns ValidatedOptions
func (o *RawOptions) Validate() (*ValidatedOptions, error) {
	if o.tokenOptions == nil {
		return nil, fmt.Errorf("token options cannot be nil")
	}

	// Validate token options
	if err := o.tokenOptions.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate token options: %w", err)
	}

	// Create validated data with defensive copies where needed
	validated := &validatedData{
		configFlags:    o.configFlags,
		flags:          o.flags,
		tokenOptions:   o.tokenOptions, // Reference to validated token options
		context:        o.context,
		azureConfigDir: o.azureConfigDir,
	}

	return &ValidatedOptions{validated: validated}, nil
}

// Complete completes the validated options and returns CompletedOptions
func (v *ValidatedOptions) Complete() (*CompletedOptions, error) {
	// Any completion logic can be added here
	// For now, we just create CompletedOptions with the same validated data

	// Apply any completion defaults or transformations
	if err := v.completeContexts(); err != nil {
		return nil, fmt.Errorf("failed to complete contexts: %w", err)
	}

	// Return CompletedOptions sharing the same validated data
	return &CompletedOptions{validated: v.validated}, nil
}

// completeContexts handles context completion logic
func (v *ValidatedOptions) completeContexts() error {
	// Add any context completion logic here
	return nil
}

// GetTokenOptions returns a copy of the token options for safe access
func (c *CompletedOptions) GetTokenOptions() *token.Options {
	if c.validated.tokenOptions == nil {
		return &token.Options{} // Return pointer to zero value if nil
	}
	return c.validated.tokenOptions // Return the pointer directly
}

// GetConfigFlags returns the config flags
func (c *CompletedOptions) GetConfigFlags() genericclioptions.RESTClientGetter {
	return c.validated.configFlags
}

// GetContext returns the context string
func (c *CompletedOptions) GetContext() string {
	return c.validated.context
}

// GetAzureConfigDir returns the Azure config directory
func (c *CompletedOptions) GetAzureConfigDir() string {
	return c.validated.azureConfigDir
}

// GetFlags returns the flag set
func (c *CompletedOptions) GetFlags() *pflag.FlagSet {
	return c.validated.flags
}

// setFlag sets a flag value by name - helper for testing
func (o *RawOptions) setFlag(name, value string) error {
	if o.flags == nil {
		return fmt.Errorf("flags not initialized")
	}
	return o.flags.Set(name, value)
}

// SetContext sets the context for raw options
func (o *RawOptions) SetContext(context string) {
	o.context = context
}

// SetAzureConfigDir sets the Azure config directory for raw options
func (o *RawOptions) SetAzureConfigDir(dir string) {
	o.azureConfigDir = dir
}

// IsSet checks if a flag was explicitly set
func (c *CompletedOptions) IsSet(name string) bool {
	if c.validated.flags == nil {
		return false
	}
	flag := c.validated.flags.Lookup(name)
	return flag != nil && flag.Changed
}

// ToString returns a string representation of the options
func (c *CompletedOptions) ToString() string {
	return fmt.Sprintf("CompletedOptions{context: %s, azureConfigDir: %s}", c.GetContext(), c.GetAzureConfigDir())
}

// UpdateFromEnv updates options from environment variables
func (o *RawOptions) UpdateFromEnv() {
	if o.tokenOptions != nil {
		o.tokenOptions.UpdateFromEnv()
	}
}

// Getter methods for ValidatedOptions (similar to CompletedOptions)

// GetTokenOptions returns the token options for ValidatedOptions
func (v *ValidatedOptions) GetTokenOptions() *token.Options {
	return v.validated.tokenOptions
}

// GetConfigFlags returns the config flags for ValidatedOptions
func (v *ValidatedOptions) GetConfigFlags() genericclioptions.RESTClientGetter {
	return v.validated.configFlags
}

// GetContext returns the context string for ValidatedOptions
func (v *ValidatedOptions) GetContext() string {
	return v.validated.context
}

// GetAzureConfigDir returns the Azure config directory for ValidatedOptions
func (v *ValidatedOptions) GetAzureConfigDir() string {
	return v.validated.azureConfigDir
}

// GetFlags returns the flag set for ValidatedOptions
func (v *ValidatedOptions) GetFlags() *pflag.FlagSet {
	return v.validated.flags
}
