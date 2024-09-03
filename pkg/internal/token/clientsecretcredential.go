package token

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
)

type ClientSecretCredential struct {
	cred *azidentity.ClientSecretCredential
}

var _ CredentialProvider = (*ClientSecretCredential)(nil)

func newClientSecretCredential(opts *Options) (CredentialProvider, error) {
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if opts.ClientSecret == "" {
		return nil, fmt.Errorf("client secret cannot be empty")
	}
	var (
		c   azidentity.Cache
		err error
	)
	if opts.UsePersistentCache {
		c, err = cache.New(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache: %s", err)
		}
	}

	cred, err := azidentity.NewClientSecretCredential(
		opts.TenantID, opts.ClientID, opts.ClientSecret,
		&azidentity.ClientSecretCredentialOptions{
			ClientOptions: azcore.ClientOptions{Cloud: opts.GetCloudConfiguration()},
			Cache:         c,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to create client secret credential: %s", err)
	}
	return &ClientSecretCredential{cred: cred}, nil
}

func (c *ClientSecretCredential) Name() string {
	return "ClientSecretCredential"
}

func (c *ClientSecretCredential) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	panic("not implemented")
}

func (c *ClientSecretCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.cred.GetToken(ctx, opts)
}

func (c *ClientSecretCredential) NeedAuthenticate() bool {
	return false
}