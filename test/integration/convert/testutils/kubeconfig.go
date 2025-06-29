//go:build integration

package testutils

import (
	"reflect"
	"strings"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// KubeconfigValidator provides utilities for validating kubeconfig conversion results
type KubeconfigValidator struct {
	t *testing.T
}

// NewKubeconfigValidator creates a new validator
func NewKubeconfigValidator(t *testing.T) *KubeconfigValidator {
	return &KubeconfigValidator{t: t}
}

// ValidateConversion checks that a kubeconfig was converted correctly
func (kv *KubeconfigValidator) ValidateConversion(originalPath, convertedPath string, expectedChanges ConversionExpectation) {
	original, err := clientcmd.LoadFromFile(originalPath)
	if err != nil {
		kv.t.Fatalf("Failed to load original kubeconfig: %v", err)
	}

	converted, err := clientcmd.LoadFromFile(convertedPath)
	if err != nil {
		kv.t.Fatalf("Failed to load converted kubeconfig: %v", err)
	}

	kv.validateConversionResult(original, converted, expectedChanges)
}

// ConversionExpectation defines what changes are expected from conversion
type ConversionExpectation struct {
	// Expected exec command and args structure
	ExpectedCommand     string
	ExpectedArgs        []string
	ExpectedInstallHint string

	// Expected authentication changes
	ExpectedLoginMethod string
	ExpectedClientID    string
	ExpectedTenantID    string
	ExpectedServerID    string

	// Validation flags
	ShouldHaveExec            bool
	ShouldRemoveAuthProvider  bool
	ShouldPreserveClusterInfo bool
}

// validateConversionResult performs the actual validation
func (kv *KubeconfigValidator) validateConversionResult(original, converted *clientcmdapi.Config, expected ConversionExpectation) {
	// Get the current user from converted config
	currentContext := converted.Contexts[converted.CurrentContext]
	if currentContext == nil {
		kv.t.Fatal("No current context found in converted kubeconfig")
	}

	currentUser := converted.AuthInfos[currentContext.AuthInfo]
	if currentUser == nil {
		kv.t.Fatal("No current user found in converted kubeconfig")
	}

	// Validate exec configuration if expected
	if expected.ShouldHaveExec {
		if currentUser.Exec == nil {
			kv.t.Fatal("Expected exec configuration but found none")
		}

		kv.validateExecConfig(currentUser.Exec, expected)
	}

	// Validate auth provider removal if expected
	if expected.ShouldRemoveAuthProvider {
		if currentUser.AuthProvider != nil {
			kv.t.Error("Expected auth provider to be removed but it still exists")
		}
	}

	// Validate cluster info preservation if expected
	if expected.ShouldPreserveClusterInfo {
		kv.validateClusterInfoPreserved(original, converted)
	}
}

// validateExecConfig validates the exec configuration structure
func (kv *KubeconfigValidator) validateExecConfig(exec *clientcmdapi.ExecConfig, expected ConversionExpectation) {
	if expected.ExpectedCommand != "" && exec.Command != expected.ExpectedCommand {
		kv.t.Errorf("Expected command %s, got %s", expected.ExpectedCommand, exec.Command)
	}
	if expected.ExpectedInstallHint != "" && !strings.Contains(exec.InstallHint, expected.ExpectedInstallHint) {
		kv.t.Errorf("Expected install hint to contain %s, got %s", expected.ExpectedInstallHint, exec.InstallHint)
	}

	// Validate specific arguments if provided
	if len(expected.ExpectedArgs) > 0 {
		kv.validateExecArgs(exec.Args, expected)
	}
}

// validateExecArgs validates that exec args contain expected values
func (kv *KubeconfigValidator) validateExecArgs(actualArgs []string, expected ConversionExpectation) {
	argsMap := make(map[string]string)

	// Parse args into key-value pairs
	for i := 0; i < len(actualArgs)-1; i += 2 {
		if strings.HasPrefix(actualArgs[i], "--") {
			argsMap[actualArgs[i]] = actualArgs[i+1]
		}
	}

	// Check expected login method
	if expected.ExpectedLoginMethod != "" {
		if loginMethod, exists := argsMap["--login"]; !exists || loginMethod != expected.ExpectedLoginMethod {
			kv.t.Errorf("Expected login method %s, got %s", expected.ExpectedLoginMethod, loginMethod)
		}
	}

	// Check expected client ID
	if expected.ExpectedClientID != "" {
		if clientID, exists := argsMap["--client-id"]; !exists || clientID != expected.ExpectedClientID {
			kv.t.Errorf("Expected client ID %s, got %s", expected.ExpectedClientID, clientID)
		}
	}

	// Check expected tenant ID
	if expected.ExpectedTenantID != "" {
		if tenantID, exists := argsMap["--tenant-id"]; !exists || tenantID != expected.ExpectedTenantID {
			kv.t.Errorf("Expected tenant ID %s, got %s", expected.ExpectedTenantID, tenantID)
		}
	}

	// Check expected server ID
	if expected.ExpectedServerID != "" {
		if serverID, exists := argsMap["--server-id"]; !exists || serverID != expected.ExpectedServerID {
			kv.t.Errorf("Expected server ID %s, got %s", expected.ExpectedServerID, serverID)
		}
	}

	// Validate all expected args are present
	for _, expectedArg := range expected.ExpectedArgs {
		found := false
		for _, actualArg := range actualArgs {
			if actualArg == expectedArg {
				found = true
				break
			}
		}
		if !found {
			kv.t.Errorf("Expected arg %s not found in %v", expectedArg, actualArgs)
		}
	}
}

// validateClusterInfoPreserved ensures cluster information is maintained
func (kv *KubeconfigValidator) validateClusterInfoPreserved(original, converted *clientcmdapi.Config) {
	if len(original.Clusters) != len(converted.Clusters) {
		kv.t.Error("Number of clusters changed during conversion")
	}

	if len(original.Contexts) != len(converted.Contexts) {
		kv.t.Error("Number of contexts changed during conversion")
	}

	// Check that cluster servers are preserved (for non-sanitized comparisons)
	for name, origCluster := range original.Clusters {
		if convCluster, exists := converted.Clusters[name]; exists {
			if origCluster.Server != convCluster.Server {
				kv.t.Logf("Cluster server changed from %s to %s (expected for sanitized tests)",
					origCluster.Server, convCluster.Server)
			}
		}
	}
}

// LogKubeconfigDiff logs the differences between two kubeconfigs for debugging
func (kv *KubeconfigValidator) LogKubeconfigDiff(originalPath, convertedPath string) {
	original, err := clientcmd.LoadFromFile(originalPath)
	if err != nil {
		kv.t.Logf("Failed to load original for diff: %v", err)
		return
	}

	converted, err := clientcmd.LoadFromFile(convertedPath)
	if err != nil {
		kv.t.Logf("Failed to load converted for diff: %v", err)
		return
	}

	kv.t.Logf("=== KUBECONFIG CONVERSION DIFF ===")
	kv.logAuthInfoDiff(original, converted)
}

// logAuthInfoDiff logs differences in auth info between original and converted
func (kv *KubeconfigValidator) logAuthInfoDiff(original, converted *clientcmdapi.Config) {
	for name, origAuth := range original.AuthInfos {
		if convAuth, exists := converted.AuthInfos[name]; exists {
			kv.t.Logf("User %s:", name)

			if !reflect.DeepEqual(origAuth.AuthProvider, convAuth.AuthProvider) {
				kv.t.Logf("  AuthProvider changed: %+v -> %+v", origAuth.AuthProvider, convAuth.AuthProvider)
			}

			if !reflect.DeepEqual(origAuth.Exec, convAuth.Exec) {
				kv.t.Logf("  Exec changed:")
				if origAuth.Exec != nil {
					kv.t.Logf("    Original: %+v", origAuth.Exec)
				}
				if convAuth.Exec != nil {
					kv.t.Logf("    Converted: %+v", convAuth.Exec)
				}
			}
		}
	}
}
