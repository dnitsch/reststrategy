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
	OAuthPassword AuthType = "OAuthPassCredentials"
	CustomToToken AuthType = "CustomToToken"
	StaticToken   AuthType = "StaticToken"
)

// +k8s:deepcopy-gen=true
type ConfigOAuth struct {
	ServerUrl               string              `yaml:"serverUrl"`
	Scopes                  []string            `yaml:"scopes"`
	EndpointParams          map[string][]string `yaml:"endpointParams"`
	OAuthSendParamsInHeader bool                `yaml:"oAuthSendParamsInHeader"`
	// for grant_type=password use these for the addition RO auth
	ResourceOwnerUser     *string `yaml:"resourceOwnerUser,omitempty"`
	ResourceOwnerPassword *string `yaml:"resourceOwnerPass,omitempty"`
}

// +k8s:deepcopy-gen=true
// customToken stores the required data to call and process custom auth Endpoints
// returning a token. the token will need to be extracted from the response
// it will then need adding to subsequent requests in the header
// under specified key and in specified format
type CustomToken struct {
	// Url to use to POST the customRequest
	AuthUrl string `yaml:"authUrl" json:"authUrl"`
	// holds the K/V credential pair. e.g.
	// email: some@one.com
	// password: pass123
	// will post this body or send in header params that payload
	CustomAuthMap KvMapVarsAny `yaml:"credential" json:"credential"`
	// whether to send the values in the header as params
	// defaults to false and map is posted in the body
	SendInHeader bool `yaml:"inHeader" json:"inHeader"`
	// JSONPAth expression to use to get the token from response
	// e.g. "$.token"
	// empty will take the entire response as the token - raw response must be string
	ResponseKey string `yaml:"responseKey" json:"responseKey" `
	// if omitted `Authorization` will be used
	// Could be X-API-Token etc..
	HeaderKey string `yaml:"headerKey" json:"headerKey"`
	// Token prefix - if omitted Bearer will be used
	// e.g. Admin ==> `Authorization: "Admin [TOKEN]"`
	TokenPrefix string `yaml:"tokenPrefix" json:"tokenPrefix"`
}

// +k8s:deepcopy-gen=true
// Auth holds the auth strategy for all Seeders
type AuthConfig struct {
	AuthStrategy AuthType     `yaml:"type" json:"type"`
	Username     string       `yaml:"username" json:"username"`
	Password     string       `yaml:"password" json:"password"`
	OAuth        *ConfigOAuth `yaml:"oauth,omitempty" json:"oauth,omitempty"`
	CustomToken  *CustomToken `yaml:"custom,omitempty" json:"custom,omitempty"`
}

// +k8s:deepcopy-gen=true
type AuthMap map[string]AuthConfig

type passwordGrantConfig struct {
	oauthPassCredsConfig *oauth2.Config
	resourceOwnerUser    string
	resourceOwnerPass    string
}

// auth holds the auth strategy for each Action
type auth struct {
	authStrategy        AuthType
	oAuthConfig         *clientcredentials.Config
	passwordGrantConfig *passwordGrantConfig
	basicAuth           *basicAuth
	customToToken       *CustomFlowAuth
	staticToken         *staticToken
	// currentToken string
}

type basicAuth struct {
	username string
	password string
}

type CustomFlowAuth struct {
	authUrl string
	// holds the K/V pairs
	customAuthMap KvMapVarsAny
	// whether to send the values in the header as params
	// defaults to false and map is posted in the body
	sendInHeader bool
	// JSONPAth expression to use to get the token from response
	// empty will take the entire response as the token - raw response must be string
	// Default "$.access_token"
	responseKey string
	// if omitted `Authorization` will be used
	// Could be X-API-Token etc..
	headerKey string
	// Token prefix - if omitted Bearer will be used
	// e.g. Admin ==> `Authorization: "Admin [TOKEN]"`
	tokenPrefix string
	// currentToken successfully retrieved and cached for re-use
	// shuold be disabled as various things will need to be implemented
	currentToken string
}

func NewCustomFlowAuth() *CustomFlowAuth {
	return &CustomFlowAuth{
		// authUrl:       CustomToken.AuthUrl,
		// customAuthMap: v.CustomToken.CustomAuthMap,
		headerKey:    "Authorization",
		tokenPrefix:  "Bearer",
		sendInHeader: false,
		responseKey:  "$.access_token",
	}
}

func (cfa *CustomFlowAuth) WithAuthUrl(v string) *CustomFlowAuth {
	cfa.authUrl = v
	return cfa
}

func (cfa *CustomFlowAuth) WithAuthMap(v KvMapVarsAny) *CustomFlowAuth {
	cfa.customAuthMap = v
	return cfa
}

// WithHeaderKey overwrites the detault `Authorization“
func (cfa *CustomFlowAuth) WithHeaderKey(v string) *CustomFlowAuth {
	cfa.headerKey = v
	return cfa
}

// WithTokenPrefix overwrites the detault `Bearer`
func (cfa *CustomFlowAuth) WithTokenPrefix(v string) *CustomFlowAuth {
	cfa.tokenPrefix = v
	return cfa
}

// WithSendInHeader sends custom request in the header
//
//	as opposed to a in url encoded form in body POST
func (cfa *CustomFlowAuth) WithSendInHeader() *CustomFlowAuth {
	cfa.sendInHeader = true
	return cfa
}

// WithResponseKey overwrites the default "$.access_token"
func (cfa *CustomFlowAuth) WithResponseKey(v string) *CustomFlowAuth {
	cfa.responseKey = v
	return cfa
}

type staticToken struct {
	headerKey   string
	staticToken string
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
		case OAuthPassword:
			c := &passwordGrantConfig{
				oauthPassCredsConfig: &oauth2.Config{
					ClientID:     v.Username,
					ClientSecret: v.Password,
					Scopes:       v.OAuth.Scopes,
					Endpoint: oauth2.Endpoint{
						AuthURL:  v.OAuth.ServerUrl,
						TokenURL: v.OAuth.ServerUrl,
					},
				},
			}
			if v.OAuth.ResourceOwnerUser != nil {
				c.resourceOwnerUser = *v.OAuth.ResourceOwnerUser
				// panic("grant type password credentials requires a resources owner username")
			}

			if v.OAuth.ResourceOwnerPassword != nil {
				c.resourceOwnerPass = *v.OAuth.ResourceOwnerPassword
				// panic("grant type password credentials requires a resources owner username")
			}
			a.passwordGrantConfig = c
			ac[k] = a
		case Basic:
			a.basicAuth = &basicAuth{username: v.Username, password: v.Password}
			ac[k] = a
		case CustomToToken:
			a.customToToken = NewCustomFlowAuth().WithAuthMap(v.CustomToken.CustomAuthMap).WithAuthUrl(v.CustomToken.AuthUrl)

			if v.CustomToken.HeaderKey != "" {
				a.customToToken.WithHeaderKey(v.CustomToken.HeaderKey)
			}
			if v.CustomToken.TokenPrefix != "" {
				a.customToToken.WithTokenPrefix(v.CustomToken.TokenPrefix)
			}
			if v.CustomToken.SendInHeader {
				a.customToToken.WithSendInHeader()
			}
			if v.CustomToken.ResponseKey != "" {
				a.customToToken.WithResponseKey(v.CustomToken.ResponseKey)
			}
			ac[k] = a
		case StaticToken:
			a.staticToken = &staticToken{headerKey: v.Username, staticToken: v.Password}
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
func (c *CustomFlowAuth) Token(ctx context.Context, client Client, log log.Loggeriface) (CustomTokenResponse, error) {

	st, err := customTokenExchange(*c, client)
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
	// TODO: currently disabled caching of custom tokens
	// if c.currentToken == "" {
	c.currentToken = token
	// }

	return CustomTokenResponse{
		TokenPrefix: c.tokenPrefix,
		TokenValue:  c.currentToken,
		HeaderKey:   c.headerKey,
	}, nil
}

func customTokenExchange(am CustomFlowAuth, client Client) ([]byte, error) {
	var body io.Reader
	tokenResp := []byte{}
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

	resp, err := client.Do(req)
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
