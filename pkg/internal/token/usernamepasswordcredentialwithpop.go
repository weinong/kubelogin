package token

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/kubelogin/pkg/internal/pop"
)

type UsernamePasswordCredentialWithPoP struct {
	popClaims map[string]string
	cloud     cloud.Configuration
	tenantID  string
	clientID  string
	username  string
	password  string
}

var _ CredentialProvider = (*UsernamePasswordCredentialWithPoP)(nil)

func newUsernamePasswordCredentialWithPoP(opts *Options) (CredentialProvider, error) {
	if opts.ClientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}
	if opts.TenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if opts.Username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if opts.Password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	popClaimsMap, err := parsePoPClaims(opts.PoPTokenClaims)
	if err != nil {
		return nil, fmt.Errorf("unable to parse PoP claims: %s", err)
	}
	if len(popClaimsMap) == 0 {
		return nil, fmt.Errorf("number of pop claims is invalid: %d", len(popClaimsMap))
	}
	return &UsernamePasswordCredentialWithPoP{
		cloud:     opts.GetCloudConfiguration(),
		tenantID:  opts.TenantID,
		clientID:  opts.ClientID,
		popClaims: popClaimsMap,
		username:  opts.Username,
		password:  opts.Password,
	}, nil
}

func (c *UsernamePasswordCredentialWithPoP) Name() string {
	return "UsernamePasswordCredentialWithPoP"
}

func (c *UsernamePasswordCredentialWithPoP) Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error) {
	panic("not implemented")
}

func (c *UsernamePasswordCredentialWithPoP) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	token, expirationTimeUnix, err := pop.AcquirePoPTokenByUsernamePassword(
		ctx,
		c.popClaims,
		opts.Scopes,
		fmt.Sprintf(authorityFormat, c.cloud.ActiveDirectoryAuthorityHost, c.tenantID),
		c.clientID,
		c.username,
		c.password,
		nil,
	)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("failed to create PoP token using username and password credential: %w", err)
	}
	return azcore.AccessToken{Token: token, ExpiresOn: time.Unix(expirationTimeUnix, 0)}, nil
}

func (c *UsernamePasswordCredentialWithPoP) NeedAuthenticate() bool {
	return false
}