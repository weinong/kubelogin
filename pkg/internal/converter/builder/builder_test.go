package builder

import (
	"strings"
	"testing"
)

func TestNewExecArgsBuilder(t *testing.T) {
	builder := NewExecArgsBuilder()

	if builder == nil {
		t.Fatal("NewExecArgsBuilder() returned nil")
	}

	args, err := builder.Build()
	if err != nil {
		t.Fatalf("New builder should not have errors: %v", err)
	}

	if len(args) != 1 || args[0] != "get-token" {
		t.Errorf("Expected [\"get-token\"], got %v", args)
	}
}

func TestAddRequiredArgument(t *testing.T) {
	tests := []struct {
		name         string
		flag         string
		value        string
		expectError  bool
		expectedArgs []string
	}{
		{
			name:         "valid required argument",
			flag:         "--client-id",
			value:        "test-client",
			expectError:  false,
			expectedArgs: []string{"get-token", "--client-id", "test-client"},
		},
		{
			name:        "empty required argument",
			flag:        "--client-id",
			value:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewExecArgsBuilder()
			builder.AddRequiredArgument(tt.flag, tt.value)

			args, err := builder.Build()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if !strings.Contains(err.Error(), "is required") {
					t.Errorf("Expected 'is required' in error, got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(args) != len(tt.expectedArgs) {
					t.Errorf("Expected %d args, got %d", len(tt.expectedArgs), len(args))
				}
				for i, expected := range tt.expectedArgs {
					if i >= len(args) || args[i] != expected {
						t.Errorf("Expected args[%d]=%s, got %s", i, expected, args[i])
					}
				}
			}
		})
	}
}

func TestAddOptionalArgument(t *testing.T) {
	tests := []struct {
		name         string
		flag         string
		value        string
		expectedArgs []string
	}{
		{
			name:         "non-empty optional argument",
			flag:         "--environment",
			value:        "AzureCloud",
			expectedArgs: []string{"get-token", "--environment", "AzureCloud"},
		},
		{
			name:         "empty optional argument",
			flag:         "--environment",
			value:        "",
			expectedArgs: []string{"get-token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewExecArgsBuilder()
			builder.AddOptionalArgument(tt.flag, tt.value)

			args, err := builder.Build()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(args) != len(tt.expectedArgs) {
				t.Errorf("Expected %d args, got %d: %v", len(tt.expectedArgs), len(args), args)
			}

			for i, expected := range tt.expectedArgs {
				if i >= len(args) || args[i] != expected {
					t.Errorf("Expected args[%d]=%s, got %s", i, expected, args[i])
				}
			}
		})
	}
}

func TestAddFlag(t *testing.T) {
	tests := []struct {
		name         string
		flag         string
		condition    bool
		expectedArgs []string
	}{
		{
			name:         "flag with true condition",
			flag:         "--legacy",
			condition:    true,
			expectedArgs: []string{"get-token", "--legacy"},
		},
		{
			name:         "flag with false condition",
			flag:         "--legacy",
			condition:    false,
			expectedArgs: []string{"get-token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewExecArgsBuilder()
			builder.AddFlag(tt.flag, tt.condition)

			args, err := builder.Build()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(args) != len(tt.expectedArgs) {
				t.Errorf("Expected %d args, got %d: %v", len(tt.expectedArgs), len(args), args)
			}

			for i, expected := range tt.expectedArgs {
				if i >= len(args) || args[i] != expected {
					t.Errorf("Expected args[%d]=%s, got %s", i, expected, args[i])
				}
			}
		})
	}
}

func TestFluentInterface(t *testing.T) {
	builder := NewExecArgsBuilder()

	args, err := builder.
		AddRequiredArgument("--server-id", "test-server").
		AddRequiredArgument("--client-id", "test-client").
		AddOptionalArgument("--environment", "AzureCloud").
		AddFlag("--legacy", true).
		Build()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := []string{
		"get-token",
		"--server-id", "test-server",
		"--client-id", "test-client",
		"--environment", "AzureCloud",
		"--legacy",
	}

	if len(args) != len(expected) {
		t.Errorf("Expected %d args, got %d: %v", len(expected), len(args), args)
	}

	for i, exp := range expected {
		if i >= len(args) || args[i] != exp {
			t.Errorf("Expected args[%d]=%s, got %s", i, exp, args[i])
		}
	}
}

func TestMustBuild(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		builder := NewExecArgsBuilder()
		builder.AddOptionalArgument("--test", "value")

		args := builder.MustBuild()
		if len(args) != 3 || args[0] != "get-token" || args[1] != "--test" || args[2] != "value" {
			t.Errorf("Unexpected args: %v", args)
		}
	})

	t.Run("panic case", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected MustBuild to panic")
			}
		}()

		builder := NewExecArgsBuilder()
		builder.AddRequiredArgument("--required", "")
		builder.MustBuild()
	})
}

func TestConvenienceBuilders(t *testing.T) {
	t.Run("required auth args", func(t *testing.T) {
		builder := NewExecArgsBuilder()

		args, err := builder.AddRequiredAuthArgs(RequiredAuthArgs{
			ServerID: "server-123",
			ClientID: "client-456",
			TenantID: "tenant-789",
		}).Build()

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []string{
			"get-token",
			"--server-id", "server-123",
			"--client-id", "client-456",
			"--tenant-id", "tenant-789",
		}

		if len(args) != len(expected) {
			t.Errorf("Expected %d args, got %d: %v", len(expected), len(args), args)
		}

		for i, exp := range expected {
			if i >= len(args) || args[i] != exp {
				t.Errorf("Expected args[%d]=%s, got %s", i, exp, args[i])
			}
		}
	})

	t.Run("interactive args", func(t *testing.T) {
		builder := NewExecArgsBuilder()

		args, err := builder.AddInteractiveArgs(InteractiveArgs{
			RedirectURL: "http://localhost:8080",
			LoginHint:   "user@example.com",
		}).Build()

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []string{
			"get-token",
			"--redirect-url", "http://localhost:8080",
			"--login-hint", "user@example.com",
		}

		if len(args) != len(expected) {
			t.Errorf("Expected %d args, got %d: %v", len(expected), len(args), args)
		}

		for i, exp := range expected {
			if i >= len(args) || args[i] != exp {
				t.Errorf("Expected args[%d]=%s, got %s", i, exp, args[i])
			}
		}
	})
}

func TestPoPTokenValidation(t *testing.T) {
	tests := []struct {
		name        string
		popArgs     PoPTokenArgs
		expectError bool
		errorMsg    string
	}{
		{
			name:        "both enabled and claims provided",
			popArgs:     PoPTokenArgs{Enabled: true, Claims: "u=/subscriptions/test"},
			expectError: false,
		},
		{
			name:        "both disabled",
			popArgs:     PoPTokenArgs{Enabled: false, Claims: ""},
			expectError: false,
		},
		{
			name:        "enabled without claims",
			popArgs:     PoPTokenArgs{Enabled: true, Claims: ""},
			expectError: true,
			errorMsg:    "--pop-claims is required when --pop-enabled is specified",
		},
		{
			name:        "claims without enabled",
			popArgs:     PoPTokenArgs{Enabled: false, Claims: "u=/subscriptions/test"},
			expectError: true,
			errorMsg:    "--pop-enabled is required when --pop-claims is specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewExecArgsBuilder()
			builder.AddPoPTokenArgs(tt.popArgs)

			_, err := builder.Build()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	builder := NewExecArgsBuilder()

	// Add multiple errors
	builder.AddRequiredArgument("--required-1", "")
	builder.AddRequiredArgument("--required-2", "")

	if !builder.HasErrors() {
		t.Error("Expected builder to have errors")
	}

	errors := builder.GetErrors()
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}

	args, err := builder.Build()
	if err == nil {
		t.Error("Expected Build() to return error")
	}

	if args != nil {
		t.Error("Expected args to be nil when there are errors")
	}

	if !strings.Contains(err.Error(), "multiple validation errors") {
		t.Errorf("Expected 'multiple validation errors' in error, got: %v", err)
	}
}

func TestCloneAndReset(t *testing.T) {
	original := NewExecArgsBuilder()
	original.AddOptionalArgument("--test", "value")
	original.AddRequiredArgument("--required", "") // This will cause an error

	// Test clone
	clone := original.Clone()
	if clone.Length() != original.Length() {
		t.Errorf("Clone should have same length as original")
	}
	if clone.HasErrors() != original.HasErrors() {
		t.Errorf("Clone should have same error state as original")
	}

	// Test reset
	original.Reset()
	if original.Length() != 1 {
		t.Errorf("Reset should leave only get-token command")
	}
	if original.HasErrors() {
		t.Error("Reset should clear errors")
	}

	// Clone should still have original state
	if !clone.HasErrors() {
		t.Error("Clone should still have errors after original was reset")
	}
}
