package token

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/testutils"
	"github.com/golang-jwt/jwt/v4"
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
			writeTokenTo:   "", // This will be set in each test case
		}
	}

	clientID := os.Getenv(testutils.ClientID)
	tenantID := os.Getenv(testutils.TenantID)

	testCases := []struct {
		name        string
		optPtr      *Options
		expectErr   bool
		expectedTyp string
	}{
		{"valid InteractiveLogin opts", newOptions(clientID, tenantID, serverID, InteractiveLogin), false, "JWT"},
		{"valid DeviceCodeLogin opts", newOptions(clientID, tenantID, serverID, DeviceCodeLogin), false, "JWT"},
		{"valid AzureCLILogin opts", newOptions(clientID, tenantID, serverID, AzureCLILogin), false, "JWT"},
		{"empty client id", newOptions("", tenantID, serverID, "interactive"), true, ""},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("\nTest Case: %s\n", tc.name)

			// Create a temporary file for writeTokenTo
			writeTokenFile, err := os.CreateTemp("", "write_token_*")
			assert.NoError(t, err)
			tc.optPtr.writeTokenTo = writeTokenFile.Name()
			defer os.Remove(tc.optPtr.writeTokenTo) // Ensure deletion after test case

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
				// Read the token from the file and parse it, test it.
				token, err := os.ReadFile(tc.optPtr.writeTokenTo)
				tokenStr := string(token)
				assert.NoError(t, err)
				assert.NotEmpty(t, tokenStr)
				claims := jwt.MapClaims{}
				parsed, _ := jwt.ParseWithClaims(tokenStr, &claims, nil)
				assert.NotNil(t, parsed)
				assert.NotNil(t, claims)

				// Test against parsed.Header["typ"] if expectedTyp is set
				if tc.expectedTyp != "" {
					assert.Equal(t, tc.expectedTyp, parsed.Header["typ"])
				}
			}
			_ = os.Remove(tc.optPtr.tokenCacheFile)
		})
	}
}
