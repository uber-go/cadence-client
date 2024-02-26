package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

type fakeOKTokenSource struct{}
type fakeFailingTokenSource struct{}

func (f *fakeOKTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "token"}, nil
}

func (f *fakeFailingTokenSource) Token() (*oauth2.Token, error) { return nil, errors.New("error") }

func TestNewOAuthAuthorizationProviderEmptyConfigFails(t *testing.T) {
	p := NewOAuthAuthorizationProvider(OAuthAuthorizerConfig{})
	_, err := p.GetAuthToken()
	assert.Error(t, err, "empty config will always fail")
}

func TestNewOAuthAuthorizationProviderExtraParams(t *testing.T) {
	p := NewOAuthAuthorizationProvider(OAuthAuthorizerConfig{
		EndpointParams: map[string]string{"test1": "test2"},
	})
	assert.NotEmpty(t, p.config.EndpointParams)
}

func TestNewOAuthAuthorizationProviderFailingTokenSourceResultsInError(t *testing.T) {
	p := &OAuthProvider{tokenSource: &fakeFailingTokenSource{}}
	token, err := p.GetAuthToken()
	assert.Nil(t, token)
	assert.Error(t, err, "error from tokensource will result in error")
}

func TestNewOAuthAuthorizationProviderTokenSourceReturnsToken(t *testing.T) {
	p := &OAuthProvider{tokenSource: &fakeOKTokenSource{}}
	token, err := p.GetAuthToken()
	assert.Equal(t, []byte("token"), token)
	assert.NoError(t, err)
}
