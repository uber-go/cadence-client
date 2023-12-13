package internal

import (
	"context"
	"fmt"
	"net/url"

	"go.uber.org/cadence/internal/common/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// OAuthAuthorizerConfig allows to configure external OAuth provider
// This is machine to machine / service to service 2-legged OAuth2 flow
type OAuthAuthorizerConfig struct {
	// ClientID to be used for acquiring token
	ClientID string `yaml:"clientID"`

	// ClientSecret to be used for acquiring token
	ClientSecret string `yaml:"clientSecret"`

	// TokenURL is the endpoint used to get token from provider
	TokenURL string `yaml:"tokenURL"`

	// Scope specifies optional requested permissions
	Scopes []string `yaml:"scopes"`

	// EndpointParams specifies additional parameters for requests to the token endpoint.
	// This needs to be provided for some OAuth providers
	EndpointParams map[string]string `yaml:"endpointParams"`
}

type AuthorizationProvider struct {
	tokenSource oauth2.TokenSource
	config      OAuthAuthorizerConfig
}

func NewOAuthAuthorizationProvider(config OAuthAuthorizerConfig) auth.AuthorizationProvider {

	oauthConfig := clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     config.TokenURL,
		Scopes:       config.Scopes,
	}

	if config.EndpointParams != nil {
		v := url.Values{}
		for name, value := range config.EndpointParams {
			v.Set(name, value)
		}
		oauthConfig.EndpointParams = v
	}

	return &AuthorizationProvider{
		tokenSource: oauthConfig.TokenSource(context.Background()),
	}
}

func (o *AuthorizationProvider) GetAuthToken() ([]byte, error) {
	token, err := o.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("token source: %w", err)
	}

	return []byte(token.AccessToken), nil
}
