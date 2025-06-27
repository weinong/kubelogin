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

// ValidatedOptions represents options that have passed validation
type ValidatedOptions struct {
	configFlags    genericclioptions.RESTClientGetter
	flags          *pflag.FlagSet
	tokenOptions   *token.Options
	context        string
	azureConfigDir string
}

// CompletedOptions represents fully completed options ready for use
type CompletedOptions struct {
	configFlags    genericclioptions.RESTClientGetter
	flags          *pflag.FlagSet
	tokenOptions   *token.Options
	context        string
	azureConfigDir string
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
	
	return &ValidatedOptions{
		configFlags:    o.configFlags,
		flags:          o.flags,
		tokenOptions:   o.tokenOptions,
		context:        o.context,
		azureConfigDir: o.azureConfigDir,
	}, nil
}

// Complete completes the validated options and returns CompletedOptions
func (v *ValidatedOptions) Complete() (*CompletedOptions, error) {
	// Here we would load any additional configuration, resolve defaults, etc.
	// For now, we'll keep it simple and just return the completed options
	
	return &CompletedOptions{
		configFlags:    v.configFlags,
		flags:          v.flags,
		tokenOptions:   v.tokenOptions,
		context:        v.context,
		azureConfigDir: v.azureConfigDir,
	}, nil
}

// GetTokenOptions returns the token options
func (c *CompletedOptions) GetTokenOptions() *token.Options {
	return c.tokenOptions
}

// GetContext returns the context
func (c *CompletedOptions) GetContext() string {
	return c.context
}

// GetAzureConfigDir returns the Azure config directory
func (c *CompletedOptions) GetAzureConfigDir() string {
	return c.azureConfigDir
}

// GetConfigFlags returns the config flags
func (c *CompletedOptions) GetConfigFlags() genericclioptions.RESTClientGetter {
	return c.configFlags
}

// GetFlags returns the flag set
func (c *CompletedOptions) GetFlags() *pflag.FlagSet {
	return c.flags
}

// IsSet checks if a flag is set
func (c *CompletedOptions) IsSet(name string) bool {
	found := false
	c.flags.Visit(func(f *pflag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// ToString returns a string representation of the options
func (c *CompletedOptions) ToString() string {
	return fmt.Sprintf("Context: %s, %s", c.context, c.tokenOptions.ToString())
}
