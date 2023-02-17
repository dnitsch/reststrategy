package seeder

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

// +kubebuilder:validation:Enum=NoAuth;BasicAuth;OAuthClientCredentials;OAuthPassCredentials;CustomToToken;StaticToken
// AuthType specifies the type of authentication to perform
// currently only a single authType per instance is allowed
type AuthType string

const (
	NoAuth        AuthType = "NoAuth"
	Basic         AuthType = "BasicAuth"
	OAuth         AuthType = "OAuthClientCredentials"
	OAuthPassword AuthType = "OAuthPassCredentials"
	CustomToToken AuthType = "CustomToToken"
	StaticToken   AuthType = "StaticToken"
)

// +k8s:deepcopy-gen=true
// +kubebuilder:object:generate=true

type ConfigOAuth struct {
	ServerUrl string   `yaml:"serverUrl" json:"serverUrl"`
	Scopes    []string `yaml:"scopes" json:"scopes"`
	// +kubebuilder:pruning:PreserveUnknownFields
	EndpointParams          map[string][]string `yaml:"endpointParams" json:"endpointParams"`
	OAuthSendParamsInHeader bool                `yaml:"oAuthSendParamsInHeader" json:"oAuthSendParamsInHeader"`
	// for grant_type=password use these for the addition RO auth
	ResourceOwnerUser     *string `yaml:"resourceOwnerUser,omitempty" json:"resourceOwnerUser,omitempty"`
	ResourceOwnerPassword *string `yaml:"resourceOwnerPass,omitempty" json:"resourceOwnerPass,omitempty"`
}

// +k8s:deepcopy-gen=true
// +kubebuilder:object:generate=true
// CustomToken stores the required data to call and process custom auth Endpoints
// returning a token. the token will need to be extracted from the response
// it will then need adding to subsequent requests in the header
// under specified key and in specified format
type CustomToken struct {
	// Url to use to POST the customRequest
	AuthUrl string `yaml:"authUrl" json:"authUrl"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	//
	// holds the K/V credential pair. e.g.
	//
	// email: some@one.com
	// password: pass123
	// id: 12312345
	//
	// will post this body or send in header params that payload
	CustomAuthMap KvMapVarsAny `yaml:"credential" json:"credential"`
	// whether to send the values in the header as params
	// defaults to false and CustomAuthMap
	// is posted in the body as json post
	SendInHeader bool `yaml:"inHeader,omitempty" json:"inHeader,omitempty"`
	// +kubebuilder:default="$.access_token"
	// JSONPath expression to use to get the token from response
	//
	// e.g. "$.token"
	//
	// empty will take the entire response as the token - raw response must be string
	ResponseKey string `yaml:"responseKey" json:"responseKey" `
	// +kubebuilder:default=Authorization
	// if omitted `Authorization` will be used
	// Could be X-API-Token etc..
	HeaderKey string `yaml:"headerKey" json:"headerKey"`
	// +kubebuilder:default=Bearer
	// Token prefix - if omitted Bearer will be used
	// e.g. Admin ==> `Authorization: "Admin [TOKEN]"`
	TokenPrefix string `yaml:"tokenPrefix" json:"tokenPrefix"`
}

// +k8s:deepcopy-gen=true
// +kubebuilder:object:generate=true
// Auth holds the auth strategy for all Seeders
type AuthConfig struct {
	Name string `yaml:"name" json:"name"`
	// AuthStrategyType must be specified
	// and conform to the type's enum
	AuthStrategy AuthType `yaml:"type" json:"type"`
	// Username must be specified with all AuthTypes
	// will be ignored for CustomToken
	// an empty string can be provided in that case
	Username string `yaml:"username" json:"username"`
	// Password will be used as client secret in the oauth flows
	// and in basic flow as well as the StaticToken value in the header.
	//
	// Can be provided in a configmanager https://github.com/dnitsch/configmanager#config-tokens format
	Password    string       `yaml:"password" json:"password"`
	OAuth       *ConfigOAuth `yaml:"oauth,omitempty" json:"oauth,omitempty"`
	CustomToken *CustomToken `yaml:"custom,omitempty" json:"custom,omitempty"`
}

// +k8s:deepcopy-gen=true
// +kubebuilder:object:generate=true
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

// WithHeaderKey overwrites the detault `Authorization`
func (cfa *CustomFlowAuth) WithHeaderKey(v string) *CustomFlowAuth {
	if v != "" {
		cfa.headerKey = v
	}
	return cfa
}

// WithTokenPrefix overwrites the detault `Bearer`
func (cfa *CustomFlowAuth) WithTokenPrefix(v string) *CustomFlowAuth {
	if v != "" {
		cfa.tokenPrefix = v
	}
	return cfa
}

// WithSendInHeader sends custom request in the header
//
//	as opposed to a in url encoded form in body POST
func (cfa *CustomFlowAuth) WithSendInHeader(v bool) *CustomFlowAuth {
	if v {
		cfa.sendInHeader = v
	}
	return cfa
}

// WithResponseKey overwrites the default "$.access_token"
func (cfa *CustomFlowAuth) WithResponseKey(v string) *CustomFlowAuth {
	if v != "" {
		cfa.responseKey = v
	}
	return cfa
}

type staticToken struct {
	headerKey   string
	staticToken string
}

type actionAuthMap map[string]auth

func NewAuth(am AuthMap) *actionAuthMap {
	return newAuth(am)
}

// NewAuthFromList same as NewAuth but accepts a list of
// auth methods
func NewAuthFromList(aml []AuthConfig) *actionAuthMap {
	am := AuthMap{}
	for _, v := range aml {
		am[v.Name] = v
	}
	return newAuth(am)
}

func newAuth(am AuthMap) *actionAuthMap {
	ac := actionAuthMap{}
	for k, v := range am {
		a := auth{}
		a.authStrategy = v.AuthStrategy
		switch strategy := v.AuthStrategy; strategy {
		case OAuth:
			a.oAuthConfig = NewClientCredentialsGrant(v)
			ac[k] = a
		case OAuthPassword:
			a.passwordGrantConfig = NewPasswordCredentialsGrant(v)
			ac[k] = a
		case Basic:
			a.basicAuth = &basicAuth{username: v.Username, password: v.Password}
			ac[k] = a
		case CustomToToken:
			a.customToToken = NewCustomFlowAuth().WithAuthMap(v.CustomToken.CustomAuthMap).
				WithAuthUrl(v.CustomToken.AuthUrl).
				WithHeaderKey(v.CustomToken.HeaderKey).
				WithResponseKey(v.CustomToken.ResponseKey).
				WithTokenPrefix(v.CustomToken.TokenPrefix).
				WithSendInHeader(v.CustomToken.SendInHeader)
			ac[k] = a
		case StaticToken:
			a.staticToken = &staticToken{headerKey: v.Username, staticToken: v.Password}
			ac[k] = a
		default:
			// will log strategy runtime error if not found
			ac[k] = auth{
				authStrategy: NoAuth,
			}
		}
	}
	return &ac
}

func NewPasswordCredentialsGrant(v AuthConfig) *passwordGrantConfig {
	pg := &passwordGrantConfig{
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
		pg.resourceOwnerUser = *v.OAuth.ResourceOwnerUser
		// panic("grant type password credentials requires a resources owner username")
	}

	if v.OAuth.ResourceOwnerPassword != nil {
		pg.resourceOwnerPass = *v.OAuth.ResourceOwnerPassword
		// panic("grant type password credentials requires a resources owner username")
	}
	return pg
}

func NewClientCredentialsGrant(v AuthConfig) *clientcredentials.Config {
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
	return c
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
	// currently disabled caching of custom tokens
	// enable retry flow by attaching a retry function
	// in the absence of formal flow to do this
	// an attempt can be made inside the first failed try on any rest call
	// and a retry can be triggered to update the token on 401/403 response
	// however it will have to assume that a token provider returns correct responses
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
