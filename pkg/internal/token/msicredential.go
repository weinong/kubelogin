package token

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type ManagedIdentityCredential struct {
	cred *azidentity.ManagedIdentityCredential
}

var _ CredentialProvider = (*ManagedIdentityCredential)(nil)

func newManagedIdentityCredential(opts *Options) (CredentialProvider, error) {
	var id azidentity.ManagedIDKind
	if opts.ClientID != "" {
		id = azidentity.ClientID(opts.ClientID)
	} else if opts.IdentityResourceID != "" {
		id = azidentity.ResourceID(opts.IdentityResourceID)
	}
	cred, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
		ClientOptions: azcore.ClientOptions{Cloud: opts.GetCloudConfiguration()},
		ID:            id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create managed identity credential: %s", err)
	}
	return &ManagedIdentityCredential{cred: cred}, nil
}

func (c *ManagedIdentityCredential) Name() string {
	return "ManagedIdentityCredential"
}

func (c *ManagedIdentityCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	panic("not implemented")
}

func (c *ManagedIdentityCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.cred.GetToken(ctx, opts)
}

func (c *ManagedIdentityCredential) NeedAuthenticate() bool {
	return false
}