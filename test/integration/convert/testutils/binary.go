//go:build integration

package testutils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// KubeloginBinary provides utilities for executing the kubelogin binary in tests
type KubeloginBinary struct {
	BinaryPath string
	t          *testing.T
}

// NewKubeloginBinary creates a new binary executor for tests
func NewKubeloginBinary(t *testing.T) *KubeloginBinary {
	// Find the kubelogin binary - should be built by make target
	binaryPath := findKubeloginBinary()
	if binaryPath == "" {
		t.Fatal("kubelogin binary not found. Run 'make kubelogin' first.")
	}

	return &KubeloginBinary{
		BinaryPath: binaryPath,
		t:          t,
	}
}

// findKubeloginBinary locates the kubelogin binary
func findKubeloginBinary() string {
	// Try common locations relative to test directory
	possiblePaths := []string{
		"../../../bin/linux_amd64/kubelogin",
		"../../../bin/darwin_amd64/kubelogin",
		"../../../bin/darwin_arm64/kubelogin",
		"../../../bin/windows_amd64/kubelogin.exe",
		"../../../kubelogin",
	}

	for _, path := range possiblePaths {
		if abs, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}

	// Try looking in PATH
	if path, err := exec.LookPath("kubelogin"); err == nil {
		return path
	}

	return ""
}

// ConvertKubeconfig executes kubelogin convert-kubeconfig with the given arguments
func (kb *KubeloginBinary) ConvertKubeconfig(kubeconfigPath string, args ...string) (string, string, error) {
	// Build command arguments
	cmdArgs := []string{"convert-kubeconfig"}
	cmdArgs = append(cmdArgs, "--kubeconfig", kubeconfigPath)
	cmdArgs = append(cmdArgs, args...)

	// Execute command
	cmd := exec.Command(kb.BinaryPath, cmdArgs...)

	// Capture both stdout and stderr
	stdout, stderr, err := runCommand(cmd)

	// Log command execution for debugging
	kb.t.Logf("Executed: %s %s", kb.BinaryPath, strings.Join(cmdArgs, " "))
	if stdout != "" {
		kb.t.Logf("STDOUT: %s", stdout)
	}
	if stderr != "" {
		kb.t.Logf("STDERR: %s", stderr)
	}
	if err != nil {
		kb.t.Logf("ERROR: %v", err)
	}

	return stdout, stderr, err
}

// ConvertKubeconfigWithFlags is a convenience method for common conversion scenarios
func (kb *KubeloginBinary) ConvertKubeconfigWithFlags(kubeconfigPath string, loginMethod string, additionalFlags map[string]string) (string, string, error) {
	args := []string{}

	if loginMethod != "" {
		args = append(args, "--login", loginMethod)
	}

	// Add additional flags
	for flag, value := range additionalFlags {
		if value != "" {
			args = append(args, fmt.Sprintf("--%s", flag), value)
		} else {
			// Boolean flag
			args = append(args, fmt.Sprintf("--%s", flag))
		}
	}

	return kb.ConvertKubeconfig(kubeconfigPath, args...)
}

// runCommand executes a command and returns stdout, stderr, and error
func runCommand(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// CreateTempKubeconfig creates a temporary kubeconfig file for testing
func CreateTempKubeconfig(t *testing.T, content string) string {
	tempDir := t.TempDir()
	kubeconfigPath := filepath.Join(tempDir, "kubeconfig.yaml")

	err := os.WriteFile(kubeconfigPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp kubeconfig: %v", err)
	}

	return kubeconfigPath
}

// CopyKubeconfigToTemp copies a kubeconfig fixture to a temporary file
func CopyKubeconfigToTemp(t *testing.T, fixturePath string) string {
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixture %s: %v", fixturePath, err)
	}

	return CreateTempKubeconfig(t, string(content))
}

// PrepareKubeconfigForTest prepares a kubeconfig file for testing, choosing the optimal destination
// If output directory is set, it creates the file there directly; otherwise uses temp directory
func PrepareKubeconfigForTest(t *testing.T, fixturePath, testName string) string {
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixture %s: %v", fixturePath, err)
	}

	outputDir := os.Getenv("KUBELOGIN_TEST_OUTPUT_DIR")
	if outputDir != "" {
		// Create output directory if it doesn't exist
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			t.Logf("Failed to create output directory %s, falling back to temp: %v", outputDir, err)
			return CreateTempKubeconfig(t, string(content))
		}

		// Create sanitized filename from test name
		sanitizedTestName := strings.ReplaceAll(testName, " ", "_")
		sanitizedTestName = strings.ReplaceAll(sanitizedTestName, "/", "_")
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.yaml", sanitizedTestName))

		// Write directly to output directory
		err = os.WriteFile(outputPath, content, 0644)
		if err != nil {
			t.Logf("Failed to write to output directory %s, falling back to temp: %v", outputPath, err)
			return CreateTempKubeconfig(t, string(content))
		}

		t.Logf("Test kubeconfig prepared at: %s", outputPath)
		return outputPath
	}

	// No output directory specified, use temp file
	return CreateTempKubeconfig(t, string(content))
}
