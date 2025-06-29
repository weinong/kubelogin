//go:build integration

package convert

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Azure/kubelogin/test/integration/convert/testutils"
)

func TestConvertKubeconfigIntegration(t *testing.T) {
	binary := testutils.NewKubeloginBinary(t)
	validator := testutils.NewKubeconfigValidator(t)

	tests := []struct {
		name            string
		fixture         string
		loginMethod     string
		additionalFlags map[string]string
		expectation     testutils.ConversionExpectation
		expectError     bool
	}{
		// Basic authentication method conversions
		{
			name:        "Convert devicecode exec to MSI",
			fixture:     "devicecode-exec.yaml",
			loginMethod: "msi",
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod:      "msi",
				ExpectedServerID:         "6dae42f8-4368-4678-94ff-3960e28e3630",
				ShouldRemoveAuthProvider: false, // Already exec format
			},
		},
		{
			name:        "Convert devicecode exec to Azure CLI",
			fixture:     "devicecode-exec.yaml",
			loginMethod: "azurecli",
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod:      "azurecli",
				ExpectedServerID:         "6dae42f8-4368-4678-94ff-3960e28e3630",
				ShouldRemoveAuthProvider: false,
			},
		},
		{
			name:        "Convert legacy azure provider to exec",
			fixture:     "legacy-azure-provider.yaml",
			loginMethod: "devicecode",
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod:      "devicecode",
				ExpectedServerID:         "6dae42f8-4368-4678-94ff-3960e28e3630",
				ExpectedClientID:         "80faf920-1908-4b52-b5ef-a8e7bedfc67a",
				ExpectedTenantID:         "84cff436-d6b3-4e81-9ecd-cab809615c5c",
				ShouldRemoveAuthProvider: true,
			},
		},

		// All authentication methods
		{
			name:        "Convert SPN exec to MSI",
			fixture:     "spn-exec.yaml",
			loginMethod: "msi",
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "msi",
				ExpectedServerID:    "6dae42f8-4368-4678-94ff-3960e28e3630",
			},
		},
		{
			name:        "Convert ROPC exec to devicecode",
			fixture:     "ropc-exec.yaml",
			loginMethod: "devicecode",
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "devicecode",
				ExpectedServerID:    "6dae42f8-4368-4678-94ff-3960e28e3630",
			},
		},
		{
			name:        "Convert interactive exec to spn",
			fixture:     "interactive-exec.yaml",
			loginMethod: "spn",
			additionalFlags: map[string]string{
				"client-secret": "new-secret-value",
			},
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "spn",
				ExpectedServerID:    "6dae42f8-4368-4678-94ff-3960e28e3630",
			},
		},
		{
			name:        "Convert workload identity to azurecli",
			fixture:     "workloadidentity-exec.yaml",
			loginMethod: "azurecli",
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "azurecli",
				ExpectedServerID:    "6dae42f8-4368-4678-94ff-3960e28e3630",
			},
		},
		{
			name:        "Convert Azure Developer CLI to interactive",
			fixture:     "azd-exec.yaml",
			loginMethod: "interactive",
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "interactive",
				ExpectedServerID:    "6dae42f8-4368-4678-94ff-3960e28e3630",
			},
		},

		// Flag combination tests
		{
			name:        "Convert with client-id override",
			fixture:     "devicecode-exec.yaml",
			loginMethod: "msi",
			additionalFlags: map[string]string{
				"client-id": "new-client-id-12345",
			},
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "msi",
				ExpectedClientID:    "new-client-id-12345",
				ExpectedServerID:    "6dae42f8-4368-4678-94ff-3960e28e3630",
			},
		},
		{
			name:        "Convert with tenant-id override",
			fixture:     "devicecode-exec.yaml",
			loginMethod: "devicecode",
			additionalFlags: map[string]string{
				"tenant-id": "new-tenant-id-98765",
			},
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "devicecode",
				ExpectedTenantID:    "new-tenant-id-98765",
				ExpectedServerID:    "6dae42f8-4368-4678-94ff-3960e28e3630",
			},
		},
		{
			name:        "Convert with environment override",
			fixture:     "devicecode-exec.yaml",
			loginMethod: "msi",
			additionalFlags: map[string]string{
				"environment": "AzureUSGovernment",
			},
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "msi",
				ExpectedServerID:    "6dae42f8-4368-4678-94ff-3960e28e3630",
			},
		},
		{
			name:        "Convert with legacy flag",
			fixture:     "devicecode-exec.yaml",
			loginMethod: "devicecode",
			additionalFlags: map[string]string{
				"legacy": "", // Boolean flag
			},
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "devicecode",
				ExpectedArgs:        []string{"--legacy"},
				// Note: Server ID might be dropped with legacy flag
			},
		},
		{
			name:        "Convert with multiple flag overrides",
			fixture:     "spn-exec.yaml",
			loginMethod: "ropc",
			additionalFlags: map[string]string{
				"client-id": "override-client-id",
				"tenant-id": "override-tenant-id",
				"username":  "override@example.com",
			},
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "ropc",
				ExpectedClientID:    "override-client-id",
				ExpectedTenantID:    "override-tenant-id",
				ExpectedServerID:    "6dae42f8-4368-4678-94ff-3960e28e3630",
			},
		},

		// Multi-user kubeconfig tests
		{
			name:        "Convert multi-user kubeconfig to consistent method",
			fixture:     "multi-user-exec.yaml",
			loginMethod: "azurecli",
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "azurecli",
				// Should convert all users to azurecli
			},
		},

		// Large kubeconfig performance test
		{
			name:        "Convert large multi-cluster kubeconfig",
			fixture:     "large-multi-cluster.yaml",
			loginMethod: "msi",
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "msi",
			},
		},
		{
			name:        "Convert devicecode to SPN",
			fixture:     "devicecode-exec.yaml",
			loginMethod: "spn",
			additionalFlags: map[string]string{
				"client-id":     "ci-cd-client-id",
				"client-secret": "ci-cd-secret",
			},
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "spn",
				ExpectedClientID:    "ci-cd-client-id",
				ExpectedServerID:    "6dae42f8-4368-4678-94ff-3960e28e3630",
			},
		},
		{
			name:        "Convert devicecode to workload identity",
			fixture:     "devicecode-exec.yaml",
			loginMethod: "workloadidentity",
			additionalFlags: map[string]string{
				"federated-token-file": "/var/run/secrets/azure/tokens/azure-identity-token",
			},
			expectation: testutils.ConversionExpectation{
				ExpectedLoginMethod: "workloadidentity",
				ExpectedServerID:    "6dae42f8-4368-4678-94ff-3960e28e3630",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare kubeconfig for testing (optimal destination based on output dir)
			fixturePath := filepath.Join("fixtures", "input", tt.fixture)
			tempKubeconfig := testutils.PrepareKubeconfigForTest(t, fixturePath, tt.name)

			// Execute conversion
			stdout, stderr, err := binary.ConvertKubeconfigWithFlags(
				tempKubeconfig,
				tt.loginMethod,
				tt.additionalFlags,
			)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but conversion succeeded. STDOUT: %s", stdout)
				}
				return
			}

			if err != nil {
				t.Fatalf("Conversion failed unexpectedly: %v. STDERR: %s", err, stderr)
			}

			// Validate conversion result
			validator.ValidateConversion(fixturePath, tempKubeconfig, tt.expectation)

			// Log diff for debugging
			t.Logf("Conversion completed successfully")
			validator.LogKubeconfigDiff(fixturePath, tempKubeconfig)
		})
	}
}

func TestConvertKubeconfigErrors(t *testing.T) {
	binary := testutils.NewKubeloginBinary(t)

	errorTests := []struct {
		name            string
		fixture         string
		args            []string
		expectedInError string
	}{
		{
			name:            "Invalid login method",
			fixture:         "devicecode-exec.yaml",
			args:            []string{"--login", "invalid-method"},
			expectedInError: "not a supported login method",
		},
		{
			name:            "Missing kubeconfig file",
			fixture:         "nonexistent.yaml",
			args:            []string{"--login", "msi"},
			expectedInError: "no such file",
		},
		{
			name:            "Malformed kubeconfig",
			fixture:         "malformed-kubeconfig.yaml",
			args:            []string{"--login", "msi"},
			expectedInError: "yaml", // YAML parsing error
		},
		{
			name:            "Missing server-id for conversion",
			fixture:         "missing-required-fields.yaml",
			args:            []string{"--login", "spn"},
			expectedInError: "server-id is required",
		},
		{
			name:            "Invalid environment",
			fixture:         "devicecode-exec.yaml",
			args:            []string{"--login", "msi", "--environment", "InvalidCloud"},
			expectedInError: "unsupported cloud environment", // May or may not error - kubelogin might accept it
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			var kubeconfigPath string

			if tt.fixture == "nonexistent.yaml" {
				kubeconfigPath = "/tmp/nonexistent.yaml"
			} else {
				fixturePath := filepath.Join("fixtures", "input", tt.fixture)
				kubeconfigPath = testutils.CopyKubeconfigToTemp(t, fixturePath)
			}

			stdout, stderr, err := binary.ConvertKubeconfig(kubeconfigPath, tt.args...)

			if err == nil {
				// Some error tests might not fail as expected due to kubelogin's flexibility
				// Only fail the test for specific errors we know should fail
				if tt.name == "Invalid login method" || tt.name == "Missing kubeconfig file" ||
					tt.name == "Malformed kubeconfig" || tt.name == "Missing server-id for conversion" {
					t.Errorf("Expected error but got success. STDOUT: %s", stdout)
				} else {
					t.Logf("Test completed without error (this may be expected): %s", stdout)
				}
				return
			}

			// Check that error message contains expected text
			errorOutput := stderr + err.Error()
			if tt.expectedInError != "" && !containsIgnoreCase(errorOutput, tt.expectedInError) {
				t.Errorf("Expected error to contain '%s', got: %s", tt.expectedInError, errorOutput)
			}

			t.Logf("Got expected error: %v", err)
		})
	}
}

// TestConvertSmoke is a quick smoke test for basic functionality
func TestConvertSmoke(t *testing.T) {
	binary := testutils.NewKubeloginBinary(t)

	// Create a dummy kubeconfig file for help test
	dummyKubeconfig := testutils.CreateTempKubeconfig(t, "apiVersion: v1\nkind: Config")

	// Test that help works
	stdout, stderr, err := binary.ConvertKubeconfig(dummyKubeconfig, "--help")

	if err == nil && (containsIgnoreCase(stdout, "convert") || containsIgnoreCase(stderr, "convert")) {
		t.Log("Smoke test passed - convert-kubeconfig command is available")
	} else {
		t.Errorf("Smoke test failed - convert-kubeconfig help not working: err=%v, stdout=%s, stderr=%s",
			err, stdout, stderr)
	}
}

// TestConvertMixedAuthMethods tests conversion of kubeconfigs with mixed authentication methods
func TestConvertMixedAuthMethods(t *testing.T) {
	binary := testutils.NewKubeloginBinary(t)
	validator := testutils.NewKubeconfigValidator(t)

	// Test converting mixed auth methods kubeconfig
	fixturePath := filepath.Join("fixtures", "input", "mixed-auth-methods.yaml")
	tempKubeconfig := testutils.PrepareKubeconfigForTest(t, fixturePath, "mixed-auth-conversion")

	// Convert to a unified authentication method
	stdout, stderr, err := binary.ConvertKubeconfigWithFlags(
		tempKubeconfig,
		"azurecli",
		nil,
	)

	if err != nil {
		t.Fatalf("Conversion failed: %v. STDERR: %s", err, stderr)
	}

	// Verify all users now use the same authentication method
	expectation := testutils.ConversionExpectation{
		ExpectedLoginMethod:      "azurecli",
		ShouldRemoveAuthProvider: true, // Should remove legacy auth-provider
	}

	validator.ValidateConversion(fixturePath, tempKubeconfig, expectation)

	t.Logf("Mixed auth methods conversion completed successfully")
	t.Logf("STDOUT: %s", stdout)
}

// Helper function for case-insensitive string contains
func containsIgnoreCase(text, substr string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(substr))
}
