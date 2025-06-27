package mapper

import (
	"testing"

	"github.com/Azure/kubelogin/pkg/internal/token"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	
	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	
	if len(registry.mappings) == 0 {
		t.Fatal("NewRegistry() should have default mappings")
	}
	
	// Verify we have expected mappings
	expectedFlags := []string{
		"client-id", "server-id", "tenant-id", "environment",
		"login-hint", "redirect-url", "client-secret", "client-certificate",
		"client-certificate-password", "username", "password",
		"identity-resource-id", "authority-host", "federated-token-file",
		"cache-dir", "pop-claims", "legacy", "pop-enabled",
		"disable-environment-override",
	}
	
	for _, flagName := range expectedFlags {
		if _, found := registry.GetMapping(flagName); !found {
			t.Errorf("Expected flag mapping for %s not found", flagName)
		}
	}
}

func TestGetMapping(t *testing.T) {
	registry := NewRegistry()
	
	tests := []struct {
		name      string
		flagName  string
		shouldFind bool
		expectedArg string
	}{
		{
			name:        "existing flag",
			flagName:    "client-id",
			shouldFind:  true,
			expectedArg: "--client-id",
		},
		{
			name:        "login-hint flag",
			flagName:    "login-hint",
			shouldFind:  true,
			expectedArg: "--login-hint",
		},
		{
			name:       "non-existing flag",
			flagName:   "non-existent",
			shouldFind: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, found := registry.GetMapping(tt.flagName)
			
			if found != tt.shouldFind {
				t.Errorf("GetMapping(%s) found=%v, want=%v", tt.flagName, found, tt.shouldFind)
			}
			
			if tt.shouldFind && mapping.ArgumentName != tt.expectedArg {
				t.Errorf("GetMapping(%s) argument=%s, want=%s", tt.flagName, mapping.ArgumentName, tt.expectedArg)
			}
		})
	}
}

func TestGetMappingsForLogin(t *testing.T) {
	registry := NewRegistry()
	
	tests := []struct {
		name        string
		loginMethod string
		expectedFlags []string
	}{
		{
			name:        "interactive login",
			loginMethod: token.InteractiveLogin,
			expectedFlags: []string{
				"client-id", "server-id", "tenant-id", "environment",
				"login-hint", "redirect-url", "cache-dir", "pop-claims",
				"pop-enabled",
			},
		},
		{
			name:        "service principal login",
			loginMethod: token.ServicePrincipalLogin,
			expectedFlags: []string{
				"client-id", "server-id", "tenant-id", "environment",
				"client-secret", "client-certificate", "client-certificate-password",
				"cache-dir", "pop-claims", "legacy", "pop-enabled",
				"disable-environment-override",
			},
		},
		{
			name:        "msi login",
			loginMethod: token.MSILogin,
			expectedFlags: []string{
				"server-id", "identity-resource-id", "cache-dir",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mappings := registry.GetMappingsForLogin(tt.loginMethod)
			
			foundFlags := make(map[string]bool)
			for _, mapping := range mappings {
				foundFlags[mapping.FlagName] = true
			}
			
			for _, expectedFlag := range tt.expectedFlags {
				if !foundFlags[expectedFlag] {
					t.Errorf("Expected flag %s for login method %s, but not found", expectedFlag, tt.loginMethod)
				}
			}
		})
	}
}

func TestGetRequiredMappings(t *testing.T) {
	registry := NewRegistry()
	
	tests := []struct {
		name        string
		loginMethod string
		expectedRequired []string
	}{
		{
			name:        "interactive login required",
			loginMethod: token.InteractiveLogin,
			expectedRequired: []string{"client-id", "server-id", "tenant-id"},
		},
		{
			name:        "msi login required",
			loginMethod: token.MSILogin,
			expectedRequired: []string{"server-id"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mappings := registry.GetRequiredMappings(tt.loginMethod)
			
			foundFlags := make(map[string]bool)
			for _, mapping := range mappings {
				foundFlags[mapping.FlagName] = true
			}
			
			if len(foundFlags) != len(tt.expectedRequired) {
				t.Errorf("Expected %d required flags for %s, got %d", 
					len(tt.expectedRequired), tt.loginMethod, len(foundFlags))
			}
			
			for _, expectedFlag := range tt.expectedRequired {
				if !foundFlags[expectedFlag] {
					t.Errorf("Expected required flag %s for login method %s, but not found", 
						expectedFlag, tt.loginMethod)
				}
			}
		})
	}
}

func TestFlagMappingValues(t *testing.T) {
	registry := NewRegistry()
	
	// Create test token options
	opts := &token.Options{
		ClientID:     "test-client-id",
		ServerID:     "test-server-id",
		TenantID:     "test-tenant-id",
		Environment:  "test-environment",
		LoginHint:    "test-user@example.com",
		RedirectURL:  "http://localhost:8080",
		IsLegacy:     true,
		IsPoPTokenEnabled: true,
	}
	
	tests := []struct {
		flagName     string
		expectedValue string
		expectedBool bool
		isBoolean    bool
	}{
		{"client-id", "test-client-id", false, false},
		{"server-id", "test-server-id", false, false},
		{"tenant-id", "test-tenant-id", false, false},
		{"environment", "test-environment", false, false},
		{"login-hint", "test-user@example.com", false, false},
		{"redirect-url", "http://localhost:8080", false, false},
		{"legacy", "", true, true},
		{"pop-enabled", "", true, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			mapping, found := registry.GetMapping(tt.flagName)
			if !found {
				t.Fatalf("Mapping for %s not found", tt.flagName)
			}
			
			if mapping.IsBoolean != tt.isBoolean {
				t.Errorf("Expected IsBoolean=%v for %s, got %v", 
					tt.isBoolean, tt.flagName, mapping.IsBoolean)
			}
			
			if tt.isBoolean {
				if mapping.GetBoolValue == nil {
					t.Fatalf("GetBoolValue should not be nil for boolean flag %s", tt.flagName)
				}
				actual := mapping.GetBoolValue(opts)
				if actual != tt.expectedBool {
					t.Errorf("Expected bool value %v for %s, got %v", 
						tt.expectedBool, tt.flagName, actual)
				}
			} else {
				if mapping.GetValue == nil {
					t.Fatalf("GetValue should not be nil for string flag %s", tt.flagName)
				}
				actual := mapping.GetValue(opts)
				if actual != tt.expectedValue {
					t.Errorf("Expected value %s for %s, got %s", 
						tt.expectedValue, tt.flagName, actual)
				}
			}
		})
	}
}

func TestContainsHelper(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "item exists",
			slice:    []string{"a", "b", "c"},
			item:     "b",
			expected: true,
		},
		{
			name:     "item does not exist",
			slice:    []string{"a", "b", "c"},
			item:     "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "a",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := contains(tt.slice, tt.item)
			if actual != tt.expected {
				t.Errorf("contains(%v, %s) = %v, want %v", 
					tt.slice, tt.item, actual, tt.expected)
			}
		})
	}
}
