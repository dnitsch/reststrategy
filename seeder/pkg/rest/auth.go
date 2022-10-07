package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/dnitsch/simplelog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// AuthType specifies the type of authentication to perform
// currently only a single authType per instance is allowed
type AuthType string

const (
	Basic         AuthType = "BasicAuth"
	OAuth         AuthType = "OAuthClientCredentials"
	CustomToToken AuthType = "CustomToToken"
)

type ConfigOAuth struct {
	ServerUrl               string              `yaml:"serverUrl"`
	Scopes                  []string            `yaml:"scopes"`
	EndpointParams          map[string][]string `yaml:"endpointParams"`
	OAuthSendParamsInHeader bool                `yaml:"oAuthSendParamsInHeader"`
}

// customToken stores the required data to call and process custom auth Endpoints
// returning a token. the token will need to be extracted from the response
// it will then need adding to subsequent requests in the header
// under specified key and in specified format
type CustomToken struct {
	// Url to use to POST the customRequest
	AuthUrl string `yaml:"authUrl"`
	// holds the K/V credential pair. e.g.
	// email: some@one.com
	// password: pass123
	// will post this body or send in header params that payload
	CustomAuthMap map[string]any `yaml:"credential"`
	// whether to send the values in the header as params
	// defaults to false and map is posted in the body
	SendInHeader bool `yaml:"inHeader"`
	// JSONPAth expression to use to get the token from response
	// e.g. "$.token"
	// empty will take the entire response as the token - raw response must be string
	ResponseKey string `yaml:"responseKey"`
	// if ommited `Authorization` will be used
	// Could be X-API-Token etc..
	HeaderKey string `yaml:"headerKey"`
	// Token prefix - if omitted Bearer will be used
	// e.g. Admin ==> `Authorization: "Admin [TOKEN]"`
	TokenPrefix string `yaml:"tokenPrefix"`
}

// Auth holds the auth strategy for all Seeders
type AuthConfig struct {
	AuthStrategy AuthType     `yaml:"type"`
	Username     string       `yaml:"username"`
	Password     string       `yaml:"password"`
	OAuth        *ConfigOAuth `yaml:"oauth,omitempty"`
	CustomToken  *CustomToken `yaml:"custom,omitempty"`
}

type AuthMap map[string]AuthConfig

// auth holds the auth strategy for each Action
type auth struct {
	authStrategy  AuthType
	oAuthConfig   *clientcredentials.Config
	basicAuth     *basicAuth
	customToToken *customToToken
	// currentToken string
}

type basicAuth struct {
	username string
	password string
}

type customToToken struct {
	authUrl string
	// holds the K/V pairs
	customAuthMap map[string]any `json:""`
	// whether to send the values in the header as params
	// defaults to false and map is posted in the body
	sendInHeader bool
	// JSONPAth expression to use to get the token from response
	// empty will take the entire response as the token - raw response must be string
	// Default "$.access_token"
	responseKey string
	// if ommited `Authorization` will be used
	// Could be X-API-Token etc..
	headerKey string
	// Token prefix - if omitted Bearer will be used
	// e.g. Admin ==> `Authorization: "Admin [TOKEN]"`
	tokenPrefix string
	// currentToken successfully retrieved and cached for re-use
	// shuold be disabled as various things will need to be implemented
	currentToken string
}

type actionAuthMap map[string]auth

func NewAuth(am *AuthMap) *actionAuthMap {
	ac := actionAuthMap{}
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
		case CustomToToken:
			a.customToToken = &customToToken{
				authUrl:       v.CustomToken.AuthUrl,
				customAuthMap: v.CustomToken.CustomAuthMap,
				headerKey:     "Authorization",
				tokenPrefix:   "Bearer",
				sendInHeader:  false,
				responseKey:   "$.access_token",
			}
			if v.CustomToken.HeaderKey != "" {
				a.customToToken.headerKey = v.CustomToken.HeaderKey
			}
			if v.CustomToken.TokenPrefix != "" {
				a.customToToken.tokenPrefix = v.CustomToken.TokenPrefix
			}
			if v.CustomToken.SendInHeader {
				a.customToToken.sendInHeader = true
			}
			if v.CustomToken.ResponseKey != "" {
				a.customToToken.responseKey = v.CustomToken.ResponseKey
			}
			ac[k] = a
		default:
			a.basicAuth = &basicAuth{username: v.Username, password: v.Password}
			ac[k] = a
		}
	}
	return &ac
}

type CustomTokenResponse struct {
	HeaderKey   string
	TokenPrefix string
	TokenValue  string
}

// NOTE: for oauth an basicAuthToToken it might make sense to build a in-memory map of tokens to strategy name
func (c *customToToken) Token(ctx context.Context, log log.Loggeriface) (CustomTokenResponse, error) {

	st, err := customTokenExchange(*c)
	if err != nil {
		return CustomTokenResponse{}, err
	}
	token, err := findPathByExpression(st, c.responseKey, log)
	if err != nil {
		return CustomTokenResponse{}, err
	}
	if token == "" {
		return CustomTokenResponse{}, fmt.Errorf("unable to retrieve and parse custom token from: %v, by pathx: %s", st, c.responseKey)
	}
	// TODO: currently disable caching of custom tokens
	// if c.currentToken == "" {
	c.currentToken = token
	// }

	return CustomTokenResponse{
		TokenPrefix: c.tokenPrefix,
		TokenValue:  c.currentToken,
		HeaderKey:   c.headerKey,
	}, nil
}

func customTokenExchange(am customToToken) ([]byte, error) {
	var body io.Reader
	tokenResp := []byte{}
	c := &http.Client{}
	// call endpoint
	b, err := json.Marshal(am.customAuthMap)
	if err != nil {
		return tokenResp, err
	}

	body = strings.NewReader(string(b))

	if am.sendInHeader {
		hvals := url.Values{}
		for k, v := range am.customAuthMap {
			hvals.Set(k, fmt.Sprintf("%v", v))
		}
		body = bytes.NewReader([]byte(hvals.Encode()))
	}

	req, err := http.NewRequestWithContext(context.TODO(), "POST", am.authUrl, body)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}

	if am.sendInHeader {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.Do(req)
	if err != nil {
		return tokenResp, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 201 {
		return tokenResp, fmt.Errorf("network call indicated a non success status code: %v", resp.StatusCode)
	}

	if resp.Body != nil {
		return io.ReadAll(resp.Body)
	}

	return tokenResp, fmt.Errorf("empty response returned")
}
