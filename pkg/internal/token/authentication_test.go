package token

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestNewInteractiveBrowserCredential(t *testing.T) {
	opts := &Options{
		ClientID:           "test-client-id",
		TenantID:           "test-tenant-id",
		UsePersistentCache: false,
	}

	record := azidentity.AuthenticationRecord{}

	cred, err := newInteractiveBrowserCredential(opts, record)
	assert.NoError(t, err)
	assert.NotNil(t, cred)
	assert.Equal(t, "InteractiveBrowserCredential", cred.Name())
}

// TestInteractiveBrowserCredential_Authenticate tests the authentication methods: Interactive, DeviceCode, and AzureCLI.
// Ensure the following environment variables are set for successful tests:
// AZURE_CLIENT_ID="80faf920-1908-4b52-b5ef-a8e7bedfc67a"
// AZURE_TENANT_ID
// AZURE_RESOURCE_ID
func TestInteractiveBrowserCredential_Authenticate(t *testing.T) {
	serverID := os.Getenv("AZURE_SERVER_ID")
	if serverID == "" {
		serverID = "6dae42f8-4368-4678-94ff-3960e28e3630" // Ref: https://github.com/Azure/kubelogin/blob/main/docs/book/src/concepts/aks.md#azure-kubernetes-service-aad-server
	}

	// Generate unique token cache paths per test run
	tokenCacheDir, err := os.MkdirTemp("", "kubelogin_cache_*")
	assert.NoError(t, err)
	defer os.RemoveAll(tokenCacheDir) // Cleanup after test

	tokenCacheFile, err := os.CreateTemp(tokenCacheDir, "kubelogin_token_*")
	assert.NoError(t, err)
	tokenCacheFilePath := tokenCacheFile.Name()
	tokenCacheFile.Close() // Close immediately as we only need the path

	// Helper function to create Options instances
	newOptions := func(clientID, tenantID, serverID, loginMethod string) *Options {
		return &Options{
			ClientID:       clientID,
			TenantID:       tenantID,
			ServerID:       serverID,
			LoginMethod:    loginMethod,
			Timeout:        70000000000,
			TokenCacheDir:  tokenCacheDir,
			tokenCacheFile: tokenCacheFilePath,
		}
	}

	clientID := os.Getenv(testutils.ClientID)
	tenantID := os.Getenv(testutils.TenantID)

	testCases := []struct {
		name      string
		optPtr    *Options
		expectErr bool
	}{
		{"valid InteractiveLogin opts", newOptions(clientID, tenantID, serverID, InteractiveLogin), false},
		{"valid DeviceCodeLogin opts", newOptions(clientID, tenantID, serverID, DeviceCodeLogin), false},
		{"valid AzureCLILogin opts", newOptions(clientID, tenantID, serverID, AzureCLILogin), false},
		{"empty client id", newOptions("", tenantID, serverID, "interactive"), true},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("\nTest Case: %s\n", tc.name)

			credential, err := New(tc.optPtr)
			assert.NoError(t, err)
			assert.NotNil(t, credential)

			err = credential.Do(ctx)
			if tc.optPtr.ClientID == "" || tc.optPtr.TenantID == "" || tc.optPtr.ServerID == "" {
				tc.expectErr = true
			}
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			_ = os.Remove(tc.optPtr.tokenCacheFile)
		})
	}
}
