package converter

import (
	"testing"

	"github.com/spf13/pflag"
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
