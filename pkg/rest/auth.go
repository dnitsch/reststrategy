package rest

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// AuthType specifies the type of authentication to perform
// currently only a single authType per instance is allowed
type AuthType string

const (
	Basic        AuthType = "BasicAuth"
	OAuth        AuthType = "OAuthClientCredentials"
	BasicToToken AuthType = "BasicToToken"
)

type ConfigOAuth struct {
	ServerUrl               string              `yaml:"serverUrl"`
	Scopes                  []string            `yaml:"scopes"`
	EndpointParams          map[string][]string `yaml:"endpointParams"`
	OAuthSendParamsInHeader bool                `yaml:"oAuthSendParamsInHeader"`
}

// Auth holds the auth strategy for all Seeders
type AuthConfig struct {
	AuthStrategy AuthType          `yaml:"type"`
	Username     string            `yaml:"username"`
	Password     string            `yaml:"password"`
	OAuth        *ConfigOAuth      `yaml:"oauth,omitempty"`
	HttpHeaders  map[string]string `yaml:"httpHeaders,omitempty"`
}

// Auth holds the auth strategy for all Seeders
type Auth struct {
	config           AuthConfig
	oAuthConfig      *clientcredentials.Config
	basicAuthToToken string
}

func NewAuth(a AuthConfig) *Auth {
	ac := &Auth{
		config: a,
	}
	if a.AuthStrategy == OAuth && a.OAuth != nil {
		c := &clientcredentials.Config{
			ClientID:       a.Username,
			ClientSecret:   a.Password,
			TokenURL:       a.OAuth.ServerUrl,
			Scopes:         a.OAuth.Scopes,
			EndpointParams: a.OAuth.EndpointParams,
			AuthStyle:      oauth2.AuthStyleInParams,
		}
		if a.OAuth.OAuthSendParamsInHeader {
			c.AuthStyle = oauth2.AuthStyleInHeader
		}
		ac.oAuthConfig = c
	}
	return ac
}

// NOTE: for oauth an basicAuthToToken it might make sense to build a in-memory map of tokens to strategy name
