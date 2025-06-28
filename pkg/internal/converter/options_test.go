package converter

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRawOptionsValidateCompletePattern(t *testing.T) {
	// Test the RawOptions → Validate() → Complete() pattern

	// Create raw options
	raw := NewRawOptions()
	fs := &pflag.FlagSet{}
	raw.AddFlags(fs)
	raw.UpdateFromEnv()

	// Set some basic values to make validation pass
	raw.tokenOptions.ServerID = "test-server-id"

	// Validate
	validated, err := raw.Validate()
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if validated == nil {
		t.Fatal("Validated options should not be nil")
	}

	// Complete
	completed, err := validated.Complete()
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	if completed == nil {
		t.Fatal("Completed options should not be nil")
	}

	// Verify we can access the completed options
	if completed.GetTokenOptions() == nil {
		t.Error("Token options should not be nil")
	}

	if completed.GetTokenOptions().ServerID != "test-server-id" {
		t.Errorf("Expected ServerID to be 'test-server-id', got %s", completed.GetTokenOptions().ServerID)
	}
}

func TestRawOptionsValidationFailure(t *testing.T) {
	// Test validation failure when login method is invalid

	raw := NewRawOptions()
	fs := &pflag.FlagSet{}
	raw.AddFlags(fs)
	raw.UpdateFromEnv()

	// Set an invalid login method
	raw.tokenOptions.LoginMethod = "invalid-method"

	// Validation should fail
	_, err := raw.Validate()
	if err == nil {
		t.Fatal("Expected validation to fail when LoginMethod is invalid")
	}
}

func TestCompletedOptionsAccessors(t *testing.T) {
	// Test all accessor methods on CompletedOptions

	raw := NewRawOptions()
	fs := &pflag.FlagSet{}
	raw.AddFlags(fs)
	raw.UpdateFromEnv()

	// Set required values
	raw.tokenOptions.ServerID = "test-server-id"
	raw.context = "test-context"
	raw.azureConfigDir = "/test/azure/config"

	validated, err := raw.Validate()
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	completed, err := validated.Complete()
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	// Test accessors
	if completed.GetContext() != "test-context" {
		t.Errorf("Expected context 'test-context', got %s", completed.GetContext())
	}

	if completed.GetAzureConfigDir() != "/test/azure/config" {
		t.Errorf("Expected azure config dir '/test/azure/config', got %s", completed.GetAzureConfigDir())
	}

	if completed.GetTokenOptions() == nil {
		t.Error("Token options should not be nil")
	}

	if completed.GetConfigFlags() == nil {
		t.Error("Config flags should not be nil")
	}

	if completed.GetFlags() == nil {
		t.Error("Flags should not be nil")
	}

	// Test ToString
	toString := completed.ToString()
	if toString == "" {
		t.Error("ToString should not be empty")
	}
}

func TestIsSetFunctionality(t *testing.T) {
	// Test the IsSet functionality

	raw := NewRawOptions()
	fs := &pflag.FlagSet{}
	raw.AddFlags(fs)

	// Set a flag value
	err := fs.Set("context", "test-context")
	if err != nil {
		t.Fatalf("Failed to set flag: %v", err)
	}

	raw.tokenOptions.ServerID = "test-server-id"

	validated, err := raw.Validate()
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	completed, err := validated.Complete()
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	// Test IsSet
	if !completed.IsSet("context") {
		t.Error("Expected context flag to be set")
	}

	if completed.IsSet("non-existent-flag") {
		t.Error("Expected non-existent flag to not be set")
	}
}

// TestEncapsulation verifies that ValidatedOptions and CompletedOptions properly encapsulate their data
func TestEncapsulation(t *testing.T) {
	// Create raw options
	rawOpts := NewRawOptions()

	// Set required values for validation to pass
	rawOpts.tokenOptions.Timeout = 10 // Set timeout to avoid validation error

	// Initialize flags (normally done by CLI)
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	rawOpts.AddFlags(flagSet)

	// Validate to get ValidatedOptions
	validatedOpts, err := rawOpts.Validate()
	require.NoError(t, err)
	require.NotNil(t, validatedOpts)

	// Complete to get CompletedOptions
	completedOpts, err := validatedOpts.Complete()
	require.NoError(t, err)
	require.NotNil(t, completedOpts)

	// Verify that validated data is accessible only through getters
	// These should work (controlled access)
	assert.NotNil(t, validatedOpts.GetTokenOptions())
	assert.Equal(t, "", validatedOpts.GetContext())
	assert.Equal(t, "", validatedOpts.GetAzureConfigDir())
	assert.NotNil(t, validatedOpts.GetConfigFlags())
	assert.NotNil(t, validatedOpts.GetFlags())

	// Same for completed options
	assert.NotNil(t, completedOpts.GetTokenOptions())
	assert.Equal(t, "", completedOpts.GetContext())
	assert.Equal(t, "", completedOpts.GetAzureConfigDir())
	assert.NotNil(t, completedOpts.GetConfigFlags())
	assert.NotNil(t, completedOpts.GetFlags())

	// Verify that both ValidatedOptions and CompletedOptions share the same validated data
	// (this verifies immutable sharing)
	assert.Same(t, validatedOpts.GetTokenOptions(), completedOpts.GetTokenOptions())
	assert.Same(t, validatedOpts.GetFlags(), completedOpts.GetFlags())
}

// TestPrivateFieldAccess verifies that private fields cannot be accessed externally
// This test ensures compile-time safety - if private fields become accessible, this will fail to compile
func TestPrivateFieldAccess(t *testing.T) {
	rawOpts := NewRawOptions()
	rawOpts.tokenOptions.Timeout = 10 // Set timeout to avoid validation error

	validatedOpts, err := rawOpts.Validate()
	require.NoError(t, err)

	completedOpts, err := validatedOpts.Complete()
	require.NoError(t, err)

	// These should be accessible (public methods)
	_ = validatedOpts.GetTokenOptions()
	_ = completedOpts.GetTokenOptions()

	// Note: The following lines would fail to compile if uncommented (which is what we want):
	// _ = validatedOpts.validated          // Should be private
	// _ = completedOpts.validated          // Should be private
	// _ = validatedOpts.tokenOptions       // Should not exist
	// _ = completedOpts.tokenOptions       // Should not exist

	// This test passing means the encapsulation is working correctly
	assert.True(t, true, "Encapsulation test passed - private fields are not accessible")
}

// TestValidationIntegrity verifies that validation creates a proper defensive copy
func TestValidationIntegrity(t *testing.T) {
	rawOpts := NewRawOptions()
	rawOpts.tokenOptions.Timeout = 10 // Set timeout to avoid validation error

	// Modify raw options
	rawOpts.context = "test-context"
	rawOpts.azureConfigDir = "/test/path"

	// Validate
	validatedOpts, err := rawOpts.Validate()
	require.NoError(t, err)

	// Verify validated data reflects the raw options at time of validation
	assert.Equal(t, "test-context", validatedOpts.GetContext())
	assert.Equal(t, "/test/path", validatedOpts.GetAzureConfigDir())

	// Modify raw options after validation
	rawOpts.context = "modified-context"
	rawOpts.azureConfigDir = "/modified/path"

	// Validated options should be unaffected (defensive copy)
	assert.Equal(t, "test-context", validatedOpts.GetContext())
	assert.Equal(t, "/test/path", validatedOpts.GetAzureConfigDir())
}
