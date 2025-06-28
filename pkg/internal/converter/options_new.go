package converter

import (
	"fmt"

	"github.com/Azure/kubelogin/pkg/internal/token"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

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

// AddFlags adds all flags to the provided FlagSet
func (o *RawOptions) AddFlags(fs *pflag.FlagSet) {
	o.flags = fs

	// Add config flags
	if cf, ok := o.configFlags.(*genericclioptions.ConfigFlags); ok {
		cf.AddFlags(fs)
	}

	// Add converter-specific flags
	fs.StringVar(&o.context, flagContext, "", "The name of the kubeconfig context to use")
	fs.StringVar(&o.azureConfigDir, flagAzureConfigDir, "", "Azure CLI config path")

	// Add token options flags
	o.tokenOptions.AddFlags(fs)
}

// UpdateFromEnv updates options from environment variables
func (o *RawOptions) UpdateFromEnv() {
	o.tokenOptions.UpdateFromEnv()
}

// Validate validates the raw options and returns ValidatedOptions if successful
func (o *RawOptions) Validate() (*ValidatedOptions, error) {
	// Validate token options first
	if err := o.tokenOptions.Validate(); err != nil {
		return nil, fmt.Errorf("token options validation failed: %w", err)
	}

	// Additional converter-specific validation can be added here
	// For now, we'll keep it simple and just validate that we have required fields

	// Create private validated data
	validated := &validatedData{
		configFlags:    o.configFlags,
		flags:          o.flags,
		tokenOptions:   o.tokenOptions,
		context:        o.context,
		azureConfigDir: o.azureConfigDir,
	}

	return &ValidatedOptions{validated: validated}, nil
}

// Complete completes the validated options and returns CompletedOptions
func (v *ValidatedOptions) Complete() (*CompletedOptions, error) {
	// Here we would load any additional configuration, resolve defaults, etc.
	// For now, we'll keep it simple and just return the completed options
	// Share the validated data immutably

	return &CompletedOptions{
		validated: v.validated, // Share the same validated data
	}, nil
}

// ValidatedOptions getter methods for controlled access to validated data

// GetTokenOptions returns the token options
func (v *ValidatedOptions) GetTokenOptions() *token.Options {
	return v.validated.tokenOptions
}

// GetContext returns the context
func (v *ValidatedOptions) GetContext() string {
	return v.validated.context
}

// GetAzureConfigDir returns the Azure config directory
func (v *ValidatedOptions) GetAzureConfigDir() string {
	return v.validated.azureConfigDir
}

// GetConfigFlags returns the config flags
func (v *ValidatedOptions) GetConfigFlags() genericclioptions.RESTClientGetter {
	return v.validated.configFlags
}

// GetFlags returns the flag set
func (v *ValidatedOptions) GetFlags() *pflag.FlagSet {
	return v.validated.flags
}

// IsSet checks if a flag is set
func (v *ValidatedOptions) IsSet(name string) bool {
	found := false
	v.validated.flags.Visit(func(f *pflag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// ToString returns a string representation of the options
func (v *ValidatedOptions) ToString() string {
	return fmt.Sprintf("Context: %s, %s", v.validated.context, v.validated.tokenOptions.ToString())
}

// CompletedOptions getter methods for controlled access to validated data

// GetTokenOptions returns the token options
func (c *CompletedOptions) GetTokenOptions() *token.Options {
	return c.validated.tokenOptions
}

// GetContext returns the context
func (c *CompletedOptions) GetContext() string {
	return c.validated.context
}

// GetAzureConfigDir returns the Azure config directory
func (c *CompletedOptions) GetAzureConfigDir() string {
	return c.validated.azureConfigDir
}

// GetConfigFlags returns the config flags
func (c *CompletedOptions) GetConfigFlags() genericclioptions.RESTClientGetter {
	return c.validated.configFlags
}

// GetFlags returns the flag set
func (c *CompletedOptions) GetFlags() *pflag.FlagSet {
	return c.validated.flags
}

// IsSet checks if a flag is set
func (c *CompletedOptions) IsSet(name string) bool {
	found := false
	c.validated.flags.Visit(func(f *pflag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// ToString returns a string representation of the options
func (c *CompletedOptions) ToString() string {
	return fmt.Sprintf("Context: %s, %s", c.validated.context, c.validated.tokenOptions.ToString())
}
