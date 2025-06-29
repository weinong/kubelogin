//go:build integration

package testutils

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// SanitizationConfig defines how to sanitize kubeconfig data
type SanitizationConfig struct {
	ClusterServer string
	ClusterName   string
	UserName      string
	ContextName   string
	// Keep original IDs for testing conversion logic
	PreserveTenantID bool
	PreserveClientID bool
	PreserveServerID bool
}

// DefaultSanitizationConfig returns a safe configuration for test fixtures
func DefaultSanitizationConfig() SanitizationConfig {
	return SanitizationConfig{
		ClusterServer:    "https://test-cluster.region.azmk8s.io:443",
		ClusterName:      "test-cluster",
		UserName:         "test-user",
		ContextName:      "test-context",
		PreserveTenantID: true, // Keep for conversion testing
		PreserveClientID: true, // Keep for conversion testing
		PreserveServerID: true, // Keep for conversion testing
	}
}

// SanitizeKubeconfig removes sensitive data while preserving structure for testing
func SanitizeKubeconfig(inputPath, outputPath string, config SanitizationConfig) error {
	// Load the kubeconfig
	kubeconfig, err := clientcmd.LoadFromFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Sanitize clusters
	for name, cluster := range kubeconfig.Clusters {
		cluster.Server = config.ClusterServer
		cluster.CertificateAuthorityData = []byte("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJlVENDQVNDZ0F3SUJBZ0lSQUt0ZXN0Y2VydGlmaWNhdGVhdXRob3JpdHlkYXRhQUFBQUFBQUFBQUFBQUFBCkFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBCkFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBCkFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBCkFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBCkFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=")
		cluster.CertificateAuthority = ""

		// Update cluster name if it's not already sanitized
		if name != config.ClusterName {
			delete(kubeconfig.Clusters, name)
			kubeconfig.Clusters[config.ClusterName] = cluster
		}
	}

	// Sanitize contexts
	newContexts := make(map[string]*clientcmdapi.Context)
	for name, context := range kubeconfig.Contexts {
		context.Cluster = config.ClusterName
		context.AuthInfo = config.UserName

		contextName := config.ContextName
		if name != contextName {
			newContexts[contextName] = context
		} else {
			newContexts[name] = context
		}
	}
	kubeconfig.Contexts = newContexts
	kubeconfig.CurrentContext = config.ContextName

	// Sanitize users - preserve auth structure but clean names
	newUsers := make(map[string]*clientcmdapi.AuthInfo)
	for _, authInfo := range kubeconfig.AuthInfos {
		// Remove any client certificates or tokens
		authInfo.ClientCertificate = ""
		authInfo.ClientCertificateData = nil
		authInfo.ClientKey = ""
		authInfo.ClientKeyData = nil
		authInfo.Token = ""
		authInfo.TokenFile = ""

		// Keep exec config intact for conversion testing
		// The tenant-id, client-id, server-id are preserved for testing conversion logic

		newUsers[config.UserName] = authInfo
	}
	kubeconfig.AuthInfos = newUsers

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write sanitized kubeconfig
	return clientcmd.WriteToFile(*kubeconfig, outputPath)
}

// LoadTestKubeconfig loads a test kubeconfig from the fixtures directory
func LoadTestKubeconfig(fixtureName string) (*clientcmdapi.Config, error) {
	fixturePath := filepath.Join("fixtures", "input", fixtureName)
	return clientcmd.LoadFromFile(fixturePath)
}

// SaveTestKubeconfig saves a kubeconfig to a temporary file for testing
func SaveTestKubeconfig(config *clientcmdapi.Config, tempDir, filename string) (string, error) {
	filepath := filepath.Join(tempDir, filename)
	return filepath, clientcmd.WriteToFile(*config, filepath)
}
