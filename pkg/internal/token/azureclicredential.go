package token

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type AzureCLICredential struct {
	cred *azidentity.AzureCLICredential
}

var _ CredentialProvider = (*AzureCLICredential)(nil)

func newAzureCLICredential(opts *Options) (CredentialProvider, error) {
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	cred, err := azidentity.NewAzureCLICredential(&azidentity.AzureCLICredentialOptions{
		TenantID: opts.TenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create azure cli credential: %s", err)
	}
	return &AzureCLICredential{cred: cred}, nil
}

func (c *AzureCLICredential) Name() string {
	return "AzureCLICredential"
}

func (c *AzureCLICredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	panic("not implemented")
}

func (c *AzureCLICredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.cred.GetToken(ctx, opts)
}

func (c *AzureCLICredential) NeedAuthenticate() bool {
	return false
}