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
	AuthStrategy AuthType     `yaml:"type"`
	Username     string       `yaml:"username"`
	Password     string       `yaml:"password"`
	OAuth        *ConfigOAuth `yaml:"oauth,omitempty"`
}

type AuthMap map[string]AuthConfig

// type Auth map[string]struct {
// 	config           AuthMap
// 	oAuthConfig      *clientcredentials.Config
// 	basicAuthToToken string
// }

// Auth holds the auth strategy for each seeder
type auth struct {
	authStrategy AuthType
	oAuthConfig  *clientcredentials.Config
	basicAuth    *basicAuth
	basicToToken *basicToToken
}

type basicAuth struct {
	username string
	password string
}

type basicToToken struct {
	username     string
	password     string
	headerKey    string
	headerValFmt string
}

type authMap map[string]auth

func NewAuth(am *AuthMap) *authMap {
	ac := authMap{}
	for k, v := range *am {
		a := auth{}
		a.authStrategy = v.AuthStrategy
		switch strategy := v.AuthStrategy; strategy {
		case OAuth:
			c := &clientcredentials.Config{
				ClientID:       v.Username,
				ClientSecret:   v.Password,
				TokenURL:       v.OAuth.ServerUrl,
				Scopes:         v.OAuth.Scopes,
				EndpointParams: v.OAuth.EndpointParams,
				AuthStyle:      oauth2.AuthStyleInParams,
			}
			if v.OAuth.OAuthSendParamsInHeader {
				c.AuthStyle = oauth2.AuthStyleInHeader
			}
			a.oAuthConfig = c
			ac[k] = a
		case Basic:
			a.basicAuth = &basicAuth{username: v.Username, password: v.Password}
			ac[k] = a
		default:
			a.basicAuth = &basicAuth{username: v.Username, password: v.Password}
			ac[k] = a
		}
	}
	return &ac
}

// NOTE: for oauth an basicAuthToToken it might make sense to build a in-memory map of tokens to strategy name
