package token

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type AzureDeveloperCLICredential struct {
	cred *azidentity.AzureDeveloperCLICredential
}

var _ CredentialProvider = (*AzureDeveloperCLICredential)(nil)

func newAzureDeveloperCLICredential(opts *Options) (CredentialProvider, error) {
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	cred, err := azidentity.NewAzureDeveloperCLICredential(&azidentity.AzureDeveloperCLICredentialOptions{
		TenantID: opts.TenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create azure developer cli credential: %s", err)
	}
	return &AzureDeveloperCLICredential{cred: cred}, nil
}

func (c *AzureDeveloperCLICredential) Name() string {
	return "AzureDeveloperCLICredential"
}

func (c *AzureDeveloperCLICredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	panic("not implemented")
}

func (c *AzureDeveloperCLICredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.cred.GetToken(ctx, opts)
}

func (c *AzureDeveloperCLICredential) NeedAuthenticate() bool {
	return false
}