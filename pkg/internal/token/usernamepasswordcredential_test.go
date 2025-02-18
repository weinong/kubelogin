package token

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/stretchr/testify/assert"
)

// TestNewUsernamePasswordCredential tests the creation of a new UsernamePasswordCredential w`ith various input scenarios.
// It verifies that the appropriate errors are returned for invalid inputs and that a valid credential is created for valid inputs.
func TestNewUsernamePasswordCredential(t *testing.T) {
	testCases := []struct {
		name          string
		clientID      string
		tenantID      string
		username      string
		password      string
		expectedError string
	}{
		{
			name:          "Empty ClientID",
			tenantID:      "test-tenant-id",
			username:      "test-username",
			password:      "test-password",
			expectedError: "client ID cannot be empty",
		},
		{
			name:          "Empty TenantID",
			clientID:      "test-client-id",
			username:      "test-username",
			password:      "test-password",
			expectedError: "tenant ID cannot be empty",
		},
		{
			name:          "Empty Username",
			clientID:      "test-client-id",
			tenantID:      "test-tenant-id",
			password:      "test-password",
			expectedError: "username cannot be empty",
		},
		{
			name:          "Empty Password",
			clientID:      "test-client-id",
			tenantID:      "test-tenant-id",
			username:      "test-username",
			expectedError: "password cannot be empty",
		},
		{
			name:     "Valid Credentials",
			clientID: "test-client-id",
			tenantID: "test-tenant-id",
			username: "test-username",
			password: "test-password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := &Options{
				ClientID:           tc.clientID,
				TenantID:           tc.tenantID,
				Username:           tc.username,
				Password:           tc.password,
				UsePersistentCache: false,
			}

			record := azidentity.AuthenticationRecord{}
			cred, err := newUsernamePasswordCredential(opts, record)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, cred)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cred)
			}
		})
	}
}

// TestUsernamePasswordCredential_Authenticate tests the Authenticate method of UsernamePasswordCredential.
// It initializes the credential with test options, verifies the credential is created without error,
// and attempts to authenticate, expecting an error and a non-nil authentication record.
func TestUsernamePasswordCredential_Authenticate(t *testing.T) {
	opts := &Options{
		ClientID:           "test-client-id",
		TenantID:           "test-tenant-id",
		Username:           "test-username",
		Password:           "test-password",
		UsePersistentCache: false,
	}

	record := azidentity.AuthenticationRecord{}
	cred, err := newUsernamePasswordCredential(opts, record)
	assert.NoError(t, err)
	assert.NotNil(t, cred)

	ctx := context.Background()
	tokenOpts := &policy.TokenRequestOptions{}
	authRecord, err := cred.Authenticate(ctx, tokenOpts)
	assert.Error(t, err)
	assert.NotNil(t, authRecord)
}
