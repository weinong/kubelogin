package handlers

import (
	"strings"
	"testing"

	"github.com/Azure/kubelogin/pkg/internal/converter/builder"
	"github.com/Azure/kubelogin/pkg/internal/converter/mapper"
	"github.com/Azure/kubelogin/pkg/internal/token"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestConversionContext(t *testing.T) {
	registry := mapper.NewRegistry()
	options := &token.Options{
		ServerID: "test-server",
		ClientID: "test-client",
		TenantID: "test-tenant",
	}
	authInfo := &api.AuthInfo{}

	ctx := &ConversionContext{
		Options:          options,
		AuthInfo:         authInfo,
		IsLegacyProvider: false,
		FlagRegistry:     registry,
		IsSet:            func(flag string) bool { return false },
	}

	if ctx.Options == nil {
		t.Error("ConversionContext Options should not be nil")
	}

	if ctx.FlagRegistry == nil {
		t.Error("ConversionContext FlagRegistry should not be nil")
	}
}

func TestValidationResult(t *testing.T) {
	tests := []struct {
		name     string
		result   ValidationResult
		expected bool
	}{
		{
			name:     "valid result",
			result:   ValidationResult{IsValid: true, Errors: nil},
			expected: true,
		},
		{
			name:     "invalid result with errors",
			result:   ValidationResult{IsValid: false, Errors: []string{"error1", "error2"}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.IsValid != tt.expected {
				t.Errorf("Expected IsValid=%v, got %v", tt.expected, tt.result.IsValid)
			}
		})
	}
}

func TestBaseHandler(t *testing.T) {
	handler := BaseHandler{
		name:          "test",
		requiredFlags: []string{"server-id", "client-id"},
		optionalFlags: []string{"environment"},
	}

	if handler.GetName() != "test" {
		t.Errorf("Expected name 'test', got '%s'", handler.GetName())
	}

	requiredFlags := handler.GetRequiredFlags()
	if len(requiredFlags) != 2 {
		t.Errorf("Expected 2 required flags, got %d", len(requiredFlags))
	}

	optionalFlags := handler.GetOptionalFlags()
	if len(optionalFlags) != 1 {
		t.Errorf("Expected 1 optional flag, got %d", len(optionalFlags))
	}
}

func TestBaseHandlerValidation(t *testing.T) {
	handler := BaseHandler{
		name:          "test",
		requiredFlags: []string{"server-id", "client-id"},
		optionalFlags: []string{"environment"},
	}

	registry := mapper.NewRegistry()

	tests := []struct {
		name     string
		options  *token.Options
		expected bool
	}{
		{
			name: "valid options",
			options: &token.Options{
				ServerID: "test-server",
				ClientID: "test-client",
			},
			expected: true,
		},
		{
			name: "missing required field",
			options: &token.Options{
				ServerID: "test-server",
				// ClientID missing
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ConversionContext{
				Options:          tt.options,
				AuthInfo:         &api.AuthInfo{},
				IsLegacyProvider: false,
				FlagRegistry:     registry,
				IsSet:            func(flag string) bool { return false },
			}

			result := handler.Validate(ctx)
			if result.IsValid != tt.expected {
				t.Errorf("Expected IsValid=%v, got %v. Errors: %v", tt.expected, result.IsValid, result.Errors)
			}
		})
	}
}

func TestInteractiveLoginHandler(t *testing.T) {
	handler := NewInteractiveLoginHandler()

	if handler.GetName() != token.InteractiveLogin {
		t.Errorf("Expected name '%s', got '%s'", token.InteractiveLogin, handler.GetName())
	}

	// Test required flags
	requiredFlags := handler.GetRequiredFlags()
	expectedRequired := []string{"server-id", "client-id", "tenant-id"}
	if len(requiredFlags) != len(expectedRequired) {
		t.Errorf("Expected %d required flags, got %d", len(expectedRequired), len(requiredFlags))
	}

	// Test optional flags
	optionalFlags := handler.GetOptionalFlags()
	expectedOptional := []string{"environment", "login-hint", "redirect-url", "pop-enabled", "pop-claims"}
	if len(optionalFlags) != len(expectedOptional) {
		t.Errorf("Expected %d optional flags, got %d", len(expectedOptional), len(optionalFlags))
	}
}

func TestInteractiveLoginHandlerValidation(t *testing.T) {
	handler := NewInteractiveLoginHandler()
	registry := mapper.NewRegistry()

	tests := []struct {
		name     string
		options  *token.Options
		expected bool
		errorMsg string
	}{
		{
			name: "valid options",
			options: &token.Options{
				ServerID: "test-server",
				ClientID: "test-client",
				TenantID: "test-tenant",
			},
			expected: true,
		},
		{
			name: "missing required field",
			options: &token.Options{
				ServerID: "test-server",
				// ClientID missing
				TenantID: "test-tenant",
			},
			expected: false,
			errorMsg: "--client-id is required",
		},
		{
			name: "pop-enabled without pop-claims",
			options: &token.Options{
				ServerID:          "test-server",
				ClientID:          "test-client",
				TenantID:          "test-tenant",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "",
			},
			expected: false,
			errorMsg: "--pop-claims is required when --pop-enabled is specified",
		},
		{
			name: "pop-claims without pop-enabled",
			options: &token.Options{
				ServerID:          "test-server",
				ClientID:          "test-client",
				TenantID:          "test-tenant",
				IsPoPTokenEnabled: false,
				PoPTokenClaims:    "u=/subscriptions/test",
			},
			expected: false,
			errorMsg: "--pop-enabled is required when --pop-claims is specified",
		},
		{
			name: "valid with pop tokens",
			options: &token.Options{
				ServerID:          "test-server",
				ClientID:          "test-client",
				TenantID:          "test-tenant",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=/subscriptions/test",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ConversionContext{
				Options:          tt.options,
				AuthInfo:         &api.AuthInfo{},
				IsLegacyProvider: false,
				FlagRegistry:     registry,
				IsSet:            func(flag string) bool { return false },
			}

			result := handler.Validate(ctx)
			if result.IsValid != tt.expected {
				t.Errorf("Expected IsValid=%v, got %v. Errors: %v", tt.expected, result.IsValid, result.Errors)
			}

			if !tt.expected && tt.errorMsg != "" {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err, tt.errorMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', got errors: %v", tt.errorMsg, result.Errors)
				}
			}
		})
	}
}

func TestInteractiveLoginHandlerBuildExecArgs(t *testing.T) {
	handler := NewInteractiveLoginHandler()
	registry := mapper.NewRegistry()

	tests := []struct {
		name          string
		options       *token.Options
		expectedLen   int
		shouldContain []string
	}{
		{
			name: "minimal interactive login",
			options: &token.Options{
				ServerID: "test-server",
				ClientID: "test-client",
				TenantID: "test-tenant",
			},
			expectedLen: 7,
			shouldContain: []string{
				"get-token",
				"--server-id", "test-server",
				"--client-id", "test-client",
				"--tenant-id", "test-tenant",
			},
		},
		{
			name: "interactive login with all options",
			options: &token.Options{
				ServerID:          "test-server",
				ClientID:          "test-client",
				TenantID:          "test-tenant",
				Environment:       "AzureCloud",
				RedirectURL:       "http://localhost:8080",
				LoginHint:         "user@example.com",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=/subscriptions/test",
			},
			expectedLen: 16,
			shouldContain: []string{
				"get-token",
				"--server-id", "test-server",
				"--client-id", "test-client",
				"--tenant-id", "test-tenant",
				"--environment", "AzureCloud",
				"--redirect-url", "http://localhost:8080",
				"--login-hint", "user@example.com",
				"--pop-enabled",
				"--pop-claims", "u=/subscriptions/test",
			},
		},
		{
			name: "interactive login with partial options",
			options: &token.Options{
				ServerID:    "test-server",
				ClientID:    "test-client",
				TenantID:    "test-tenant",
				Environment: "AzureCloud",
				LoginHint:   "user@example.com",
				// No PoP tokens, no redirect URL
			},
			expectedLen: 11,
			shouldContain: []string{
				"get-token",
				"--server-id", "test-server",
				"--client-id", "test-client",
				"--tenant-id", "test-tenant",
				"--environment", "AzureCloud",
				"--login-hint", "user@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ConversionContext{
				Options:          tt.options,
				AuthInfo:         &api.AuthInfo{},
				IsLegacyProvider: false,
				FlagRegistry:     registry,
				IsSet:            func(flag string) bool { return false },
			}

			argBuilder := builder.NewExecArgsBuilder()
			err := handler.BuildExecArgs(ctx, argBuilder)

			if err != nil {
				t.Errorf("BuildExecArgs returned unexpected error: %v", err)
			}

			args, err := argBuilder.Build()
			if err != nil {
				t.Errorf("Builder.Build() returned unexpected error: %v", err)
			}

			if len(args) != tt.expectedLen {
				t.Errorf("Expected %d arguments, got %d: %v", tt.expectedLen, len(args), args)
			}

			// Check that all expected strings are present
			argsStr := strings.Join(args, " ")
			for _, expected := range tt.shouldContain {
				if !strings.Contains(argsStr, expected) {
					t.Errorf("Expected argument list to contain '%s', got: %v", expected, args)
				}
			}
		})
	}
}

func TestInteractiveLoginHandlerBuildExecArgsValidation(t *testing.T) {
	handler := NewInteractiveLoginHandler()
	registry := mapper.NewRegistry()

	// Test that PoP validation happens in the builder
	options := &token.Options{
		ServerID:          "test-server",
		ClientID:          "test-client",
		TenantID:          "test-tenant",
		IsPoPTokenEnabled: true,
		PoPTokenClaims:    "", // Invalid: enabled without claims
	}

	ctx := &ConversionContext{
		Options:          options,
		AuthInfo:         &api.AuthInfo{},
		IsLegacyProvider: false,
		FlagRegistry:     registry,
		IsSet:            func(flag string) bool { return false },
	}

	argBuilder := builder.NewExecArgsBuilder()
	err := handler.BuildExecArgs(ctx, argBuilder)

	if err != nil {
		t.Errorf("BuildExecArgs returned unexpected error: %v", err)
	}

	// The validation error should occur when building
	_, err = argBuilder.Build()
	if err == nil {
		t.Error("Expected Build() to return error for invalid PoP token configuration")
	}

	if !strings.Contains(err.Error(), "pop-claims") {
		t.Errorf("Expected error about pop-claims, got: %v", err)
	}
}

func TestDeviceCodeLoginHandler(t *testing.T) {
	handler := NewDeviceCodeLoginHandler()

	if handler.GetName() != token.DeviceCodeLogin {
		t.Errorf("Expected name '%s', got '%s'", token.DeviceCodeLogin, handler.GetName())
	}

	// Test that required flags include the core auth flags
	requiredFlags := handler.GetRequiredFlags()
	expectedRequired := []string{"server-id", "client-id", "tenant-id"}
	if len(requiredFlags) != len(expectedRequired) {
		t.Errorf("Expected %d required flags, got %d", len(expectedRequired), len(requiredFlags))
	}

	// Test that optional flags include environment and legacy
	optionalFlags := handler.GetOptionalFlags()
	expectedOptional := []string{"environment", "legacy"}
	if len(optionalFlags) != len(expectedOptional) {
		t.Errorf("Expected %d optional flags, got %d", len(expectedOptional), len(optionalFlags))
	}
}

func TestDeviceCodeLoginHandlerValidation(t *testing.T) {
	handler := NewDeviceCodeLoginHandler()
	registry := mapper.NewRegistry()

	tests := []struct {
		name     string
		options  *token.Options
		expected bool
		errorMsg string
	}{
		{
			name: "valid options",
			options: &token.Options{
				ServerID: "test-server",
				ClientID: "test-client",
				TenantID: "test-tenant",
			},
			expected: true,
		},
		{
			name: "missing client-id",
			options: &token.Options{
				ServerID: "test-server",
				// ClientID missing
				TenantID: "test-tenant",
			},
			expected: false,
			errorMsg: "--client-id is required",
		},
		{
			name: "missing tenant-id",
			options: &token.Options{
				ServerID: "test-server",
				ClientID: "test-client",
				// TenantID missing
			},
			expected: false,
			errorMsg: "--tenant-id is required",
		},
		{
			name: "valid with optional fields",
			options: &token.Options{
				ServerID:    "test-server",
				ClientID:    "test-client",
				TenantID:    "test-tenant",
				Environment: "AzureCloud",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ConversionContext{
				Options:          tt.options,
				AuthInfo:         &api.AuthInfo{},
				IsLegacyProvider: false,
				FlagRegistry:     registry,
				IsSet:            func(flag string) bool { return false },
			}

			result := handler.Validate(ctx)
			if result.IsValid != tt.expected {
				t.Errorf("Expected IsValid=%v, got %v. Errors: %v", tt.expected, result.IsValid, result.Errors)
			}

			if !tt.expected && tt.errorMsg != "" {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err, tt.errorMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', got errors: %v", tt.errorMsg, result.Errors)
				}
			}
		})
	}
}

func TestDeviceCodeLoginHandlerBuildExecArgs(t *testing.T) {
	handler := NewDeviceCodeLoginHandler()
	registry := mapper.NewRegistry()

	tests := []struct {
		name          string
		options       *token.Options
		isLegacy      bool
		expectedLen   int
		shouldContain []string
	}{
		{
			name: "minimal device code login",
			options: &token.Options{
				ServerID: "test-server",
				ClientID: "test-client",
				TenantID: "test-tenant",
			},
			isLegacy:    false,
			expectedLen: 7,
			shouldContain: []string{
				"get-token",
				"--server-id", "test-server",
				"--client-id", "test-client",
				"--tenant-id", "test-tenant",
			},
		},
		{
			name: "device code login with environment",
			options: &token.Options{
				ServerID:    "test-server",
				ClientID:    "test-client",
				TenantID:    "test-tenant",
				Environment: "AzureCloud",
			},
			isLegacy:    false,
			expectedLen: 9,
			shouldContain: []string{
				"get-token",
				"--server-id", "test-server",
				"--client-id", "test-client",
				"--tenant-id", "test-tenant",
				"--environment", "AzureCloud",
			},
		},
		{
			name: "device code login with legacy flag",
			options: &token.Options{
				ServerID: "test-server",
				ClientID: "test-client",
				TenantID: "test-tenant",
			},
			isLegacy:    true,
			expectedLen: 8,
			shouldContain: []string{
				"get-token",
				"--server-id", "test-server",
				"--client-id", "test-client",
				"--tenant-id", "test-tenant",
				"--legacy",
			},
		},
		{
			name: "device code login with all options",
			options: &token.Options{
				ServerID:    "test-server",
				ClientID:    "test-client",
				TenantID:    "test-tenant",
				Environment: "AzureCloud",
			},
			isLegacy:    true,
			expectedLen: 10,
			shouldContain: []string{
				"get-token",
				"--server-id", "test-server",
				"--client-id", "test-client",
				"--tenant-id", "test-tenant",
				"--environment", "AzureCloud",
				"--legacy",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ConversionContext{
				Options:          tt.options,
				AuthInfo:         &api.AuthInfo{},
				IsLegacyProvider: tt.isLegacy,
				FlagRegistry:     registry,
				IsSet:            func(flag string) bool { return false },
			}

			argBuilder := builder.NewExecArgsBuilder()
			err := handler.BuildExecArgs(ctx, argBuilder)

			if err != nil {
				t.Errorf("BuildExecArgs returned unexpected error: %v", err)
			}

			args, err := argBuilder.Build()
			if err != nil {
				t.Errorf("Builder.Build() returned unexpected error: %v", err)
			}

			if len(args) != tt.expectedLen {
				t.Errorf("Expected %d arguments, got %d: %v", tt.expectedLen, len(args), args)
			}

			// Check that all expected strings are present
			argsStr := strings.Join(args, " ")
			for _, expected := range tt.shouldContain {
				if !strings.Contains(argsStr, expected) {
					t.Errorf("Expected argument list to contain '%s', got: %v", expected, args)
				}
			}
		})
	}
}

func TestServicePrincipalLoginHandler(t *testing.T) {
	handler := NewServicePrincipalLoginHandler()
	
	if handler.GetName() != token.ServicePrincipalLogin {
		t.Errorf("Expected name '%s', got '%s'", token.ServicePrincipalLogin, handler.GetName())
	}
	
	// Test that required flags include the core auth flags
	requiredFlags := handler.GetRequiredFlags()
	expectedRequired := []string{"server-id", "client-id", "tenant-id"}
	if len(requiredFlags) != len(expectedRequired) {
		t.Errorf("Expected %d required flags, got %d", len(expectedRequired), len(requiredFlags))
	}
	
	// Test that optional flags include certificate and secret options
	optionalFlags := handler.GetOptionalFlags()
	if len(optionalFlags) < 3 {
		t.Errorf("Expected at least 3 optional flags, got %d", len(optionalFlags))
	}
}

func TestServicePrincipalLoginHandlerValidation(t *testing.T) {
	handler := NewServicePrincipalLoginHandler()
	registry := mapper.NewRegistry()
	
	tests := []struct {
		name     string
		options  *token.Options
		expected bool
		errorMsg string
	}{
		{
			name: "valid with client secret",
			options: &token.Options{
				ServerID:     "test-server",
				ClientID:     "test-client",
				TenantID:     "test-tenant",
				ClientSecret: "test-secret",
			},
			expected: true,
		},
		{
			name: "valid with client certificate",
			options: &token.Options{
				ServerID:   "test-server",
				ClientID:   "test-client",
				TenantID:   "test-tenant",
				ClientCert: "/path/to/cert.pem",
			},
			expected: true,
		},
		{
			name: "valid with client certificate and password",
			options: &token.Options{
				ServerID:           "test-server",
				ClientID:           "test-client",
				TenantID:           "test-tenant",
				ClientCert:         "/path/to/cert.pfx",
				ClientCertPassword: "cert-password",
			},
			expected: true,
		},
		{
			name: "missing required field",
			options: &token.Options{
				ServerID:     "test-server",
				// ClientID missing
				TenantID:     "test-tenant",
				ClientSecret: "test-secret",
			},
			expected: false,
			errorMsg: "--client-id is required",
		},
		{
			name: "missing authentication method",
			options: &token.Options{
				ServerID: "test-server",
				ClientID: "test-client",
				TenantID: "test-tenant",
				// No ClientSecret or ClientCert
			},
			expected: false,
			errorMsg: "requires either --client-secret or --client-certificate",
		},
		{
			name: "pop-enabled without pop-claims",
			options: &token.Options{
				ServerID:          "test-server",
				ClientID:          "test-client",
				TenantID:          "test-tenant",
				ClientSecret:      "test-secret",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "",
			},
			expected: false,
			errorMsg: "--pop-claims is required when --pop-enabled is specified",
		},
		{
			name: "pop-claims without pop-enabled",
			options: &token.Options{
				ServerID:          "test-server",
				ClientID:          "test-client",
				TenantID:          "test-tenant",
				ClientSecret:      "test-secret",
				IsPoPTokenEnabled: false,
				PoPTokenClaims:    "u=/subscriptions/test",
			},
			expected: false,
			errorMsg: "--pop-enabled is required when --pop-claims is specified",
		},
		{
			name: "valid with pop tokens",
			options: &token.Options{
				ServerID:          "test-server",
				ClientID:          "test-client",
				TenantID:          "test-tenant",
				ClientSecret:      "test-secret",
				IsPoPTokenEnabled: true,
				PoPTokenClaims:    "u=/subscriptions/test",
			},
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ConversionContext{
				Options:          tt.options,
				AuthInfo:         &api.AuthInfo{},
				IsLegacyProvider: false,
				FlagRegistry:     registry,
				IsSet:            func(flag string) bool { return false },
			}
			
			result := handler.Validate(ctx)
			if result.IsValid != tt.expected {
				t.Errorf("Expected IsValid=%v, got %v. Errors: %v", tt.expected, result.IsValid, result.Errors)
			}
			
			if !tt.expected && tt.errorMsg != "" {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err, tt.errorMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', got errors: %v", tt.errorMsg, result.Errors)
				}
			}
		})
	}
}

func TestServicePrincipalLoginHandlerBuildExecArgs(t *testing.T) {
	handler := NewServicePrincipalLoginHandler()
	registry := mapper.NewRegistry()
	
	tests := []struct {
		name          string
		options       *token.Options
		isLegacy      bool
		expectedLen   int
		shouldContain []string
	}{
		{
			name: "service principal with client secret",
			options: &token.Options{
				ServerID:     "test-server",
				ClientID:     "test-client",
				TenantID:     "test-tenant",
				ClientSecret: "test-secret",
			},
			isLegacy:    false,
			expectedLen: 9,
			shouldContain: []string{
				"get-token",
				"--server-id", "test-server",
				"--client-id", "test-client",
				"--tenant-id", "test-tenant",
				"--client-secret", "test-secret",
			},
		},
		{
			name: "service principal with client certificate",
			options: &token.Options{
				ServerID:   "test-server",
				ClientID:   "test-client",
				TenantID:   "test-tenant",
				ClientCert: "/path/to/cert.pem",
			},
			isLegacy:    false,
			expectedLen: 9,
			shouldContain: []string{
				"get-token",
				"--server-id", "test-server",
				"--client-id", "test-client",
				"--tenant-id", "test-tenant",
				"--client-certificate", "/path/to/cert.pem",
			},
		},
		{
			name: "service principal with certificate and password",
			options: &token.Options{
				ServerID:           "test-server",
				ClientID:           "test-client",
				TenantID:           "test-tenant",
				ClientCert:         "/path/to/cert.pfx",
				ClientCertPassword: "cert-password",
			},
			isLegacy:    false,
			expectedLen: 11,
			shouldContain: []string{
				"get-token",
				"--server-id", "test-server",
				"--client-id", "test-client",
				"--tenant-id", "test-tenant",
				"--client-certificate", "/path/to/cert.pfx",
				"--client-certificate-password", "cert-password",
			},
		},
		{
			name: "service principal with all options",
			options: &token.Options{
				ServerID:                   "test-server",
				ClientID:                   "test-client",
				TenantID:                   "test-tenant",
				Environment:                "AzureCloud",
				ClientSecret:               "test-secret",
				IsPoPTokenEnabled:          true,
				PoPTokenClaims:             "u=/subscriptions/test",
				DisableEnvironmentOverride: true,
			},
			isLegacy:    true,
			expectedLen: 16,
			shouldContain: []string{
				"get-token",
				"--server-id", "test-server",
				"--client-id", "test-client",
				"--tenant-id", "test-tenant",
				"--environment", "AzureCloud",
				"--client-secret", "test-secret",
				"--pop-enabled",
				"--pop-claims", "u=/subscriptions/test",
				"--legacy",
				"--disable-environment-override",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ConversionContext{
				Options:          tt.options,
				AuthInfo:         &api.AuthInfo{},
				IsLegacyProvider: tt.isLegacy,
				FlagRegistry:     registry,
				IsSet:            func(flag string) bool { return false },
			}
			
			argBuilder := builder.NewExecArgsBuilder()
			err := handler.BuildExecArgs(ctx, argBuilder)
			
			if err != nil {
				t.Errorf("BuildExecArgs returned unexpected error: %v", err)
			}
			
			args, err := argBuilder.Build()
			if err != nil {
				t.Errorf("Builder.Build() returned unexpected error: %v", err)
			}
			
			if len(args) != tt.expectedLen {
				t.Errorf("Expected %d arguments, got %d: %v", tt.expectedLen, len(args), args)
			}
			
			// Check that all expected strings are present
			argsStr := strings.Join(args, " ")
			for _, expected := range tt.shouldContain {
				if !strings.Contains(argsStr, expected) {
					t.Errorf("Expected argument list to contain '%s', got: %v", expected, args)
				}
			}
		})
	}
}

func TestMSILoginHandler(t *testing.T) {
	handler := NewMSILoginHandler()

	if handler.GetName() != token.MSILogin {
		t.Errorf("Expected name '%s', got '%s'", token.MSILogin, handler.GetName())
	}

	// MSI only requires server-id
	requiredFlags := handler.GetRequiredFlags()
	expectedRequired := []string{"server-id"}
	if len(requiredFlags) != len(expectedRequired) {
		t.Errorf("Expected %d required flags, got %d", len(expectedRequired), len(requiredFlags))
	}

	// Test that optional flags include client-id and identity-resource-id
	optionalFlags := handler.GetOptionalFlags()
	expectedOptional := []string{"client-id", "identity-resource-id"}
	if len(optionalFlags) != len(expectedOptional) {
		t.Errorf("Expected %d optional flags, got %d", len(expectedOptional), len(optionalFlags))
	}
}

func TestMSILoginHandlerValidation(t *testing.T) {
	handler := NewMSILoginHandler()
	registry := mapper.NewRegistry()

	tests := []struct {
		name     string
		options  *token.Options
		expected bool
		errors   []string
	}{
		{
			name: "valid MSI with server-id only",
			options: &token.Options{
				ServerID: "test-server",
			},
			expected: true,
			errors:   nil,
		},
		{
			name: "valid MSI with client-id",
			options: &token.Options{
				ServerID: "test-server",
				ClientID: "test-client",
			},
			expected: true,
			errors:   nil,
		},
		{
			name: "valid MSI with identity-resource-id",
			options: &token.Options{
				ServerID:           "test-server",
				IdentityResourceID: "test-identity-resource",
			},
			expected: true,
			errors:   nil,
		},
		{
			name: "invalid - missing server-id",
			options: &token.Options{
				ClientID: "test-client",
			},
			expected: false,
			errors:   []string{"--server-id is required"},
		},
		{
			name: "invalid - both client-id and identity-resource-id",
			options: &token.Options{
				ServerID:           "test-server",
				ClientID:           "test-client",
				IdentityResourceID: "test-identity-resource",
			},
			expected: false,
			errors:   []string{"MSI login cannot specify both --client-id and --identity-resource-id, they are mutually exclusive"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ConversionContext{
				Options:      tt.options,
				AuthInfo:     &api.AuthInfo{},
				FlagRegistry: registry,
				IsSet:        func(flag string) bool { return false },
			}

			result := handler.Validate(ctx)
			if result.IsValid != tt.expected {
				t.Errorf("Expected IsValid: %v, got: %v", tt.expected, result.IsValid)
			}

			if len(result.Errors) != len(tt.errors) {
				t.Errorf("Expected %d errors, got %d", len(tt.errors), len(result.Errors))
			}

			for i, expectedError := range tt.errors {
				if i < len(result.Errors) && !strings.Contains(result.Errors[i], expectedError) {
					t.Errorf("Expected error to contain '%s', got '%s'", expectedError, result.Errors[i])
				}
			}
		})
	}
}

func TestMSILoginHandlerBuildExecArgs(t *testing.T) {
	handler := NewMSILoginHandler()
	registry := mapper.NewRegistry()

	tests := []struct {
		name         string
		options      *token.Options
		expectedArgs []string
	}{
		{
			name: "MSI with server-id only",
			options: &token.Options{
				ServerID: "test-server",
			},
			expectedArgs: []string{"get-token", "--server-id", "test-server"},
		},
		{
			name: "MSI with client-id",
			options: &token.Options{
				ServerID: "test-server",
				ClientID: "test-client",
			},
			expectedArgs: []string{"get-token", "--server-id", "test-server", "--client-id", "test-client"},
		},
		{
			name: "MSI with identity-resource-id",
			options: &token.Options{
				ServerID:           "test-server",
				IdentityResourceID: "test-identity-resource",
			},
			expectedArgs: []string{"get-token", "--server-id", "test-server", "--identity-resource-id", "test-identity-resource"},
		},
		{
			name: "MSI with empty optional fields",
			options: &token.Options{
				ServerID:           "test-server",
				ClientID:           "", // Empty should be skipped
				IdentityResourceID: "", // Empty should be skipped
			},
			expectedArgs: []string{"get-token", "--server-id", "test-server"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ConversionContext{
				Options:      tt.options,
				AuthInfo:     &api.AuthInfo{},
				FlagRegistry: registry,
				IsSet:        func(flag string) bool { return false },
			}

			argBuilder := builder.NewExecArgsBuilder()
			err := handler.BuildExecArgs(ctx, argBuilder)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			args, buildErr := argBuilder.Build()
			if buildErr != nil {
				t.Errorf("Unexpected build error: %v", buildErr)
			}

			if len(args) != len(tt.expectedArgs) {
				t.Errorf("Expected %d args, got %d. Expected: %v, Got: %v", len(tt.expectedArgs), len(args), tt.expectedArgs, args)
			}

			for i, expectedArg := range tt.expectedArgs {
				if i < len(args) && args[i] != expectedArg {
					t.Errorf("Expected arg[%d] to be '%s', got '%s'", i, expectedArg, args[i])
				}
			}
		})
	}
}

func TestHandlerRegistry(t *testing.T) {
	registry := NewHandlerRegistry()

	// Test that default handlers are registered
	expectedHandlers := []string{
		token.InteractiveLogin,
		token.DeviceCodeLogin,
		token.ServicePrincipalLogin,
		token.MSILogin,
	}

	for _, loginMethod := range expectedHandlers {
		handler, exists := registry.GetHandler(loginMethod)
		if !exists {
			t.Errorf("Expected handler for '%s' to be registered", loginMethod)
		}
		if handler.GetName() != loginMethod {
			t.Errorf("Expected handler name '%s', got '%s'", loginMethod, handler.GetName())
		}
	}

	// Test GetAllHandlers
	allHandlers := registry.GetAllHandlers()
	if len(allHandlers) != len(expectedHandlers) {
		t.Errorf("Expected %d handlers, got %d", len(expectedHandlers), len(allHandlers))
	}
}

func TestHandlerRegistryCustomHandler(t *testing.T) {
	registry := NewHandlerRegistry()

	// Create a custom handler using an existing implementation
	customHandler := NewInteractiveLoginHandler()
	customHandler.BaseHandler.name = "custom"

	// Register it
	registry.Register(customHandler)

	// Test retrieval
	handler, exists := registry.GetHandler("custom")
	if !exists {
		t.Error("Expected custom handler to be registered")
	}
	if handler.GetName() != "custom" {
		t.Errorf("Expected handler name 'custom', got '%s'", handler.GetName())
	}
}

func TestHandlerRegistryNonExistentHandler(t *testing.T) {
	registry := NewHandlerRegistry()

	handler, exists := registry.GetHandler("non-existent")
	if exists {
		t.Error("Expected non-existent handler to not be found")
	}
	if handler != nil {
		t.Error("Expected handler to be nil for non-existent handler")
	}
}
