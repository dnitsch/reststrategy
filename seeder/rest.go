package seeder

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/a8m/envsubst"

	log "github.com/dnitsch/simplelog"
	"github.com/spyzhov/ajson"
	"k8s.io/apimachinery/pkg/runtime"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
	// RoundTrip(*http.Request) (*http.Response, error)
	// http.RoundTripper
}

type RuntimeVars struct {
	mu   sync.RWMutex
	vars map[string]any
}

type SeederImpl struct {
	log         log.Loggeriface
	client      Client
	auth        *actionAuthMap
	runtimeVars RuntimeVars
}

func NewSeederImpl(log log.Loggeriface) *SeederImpl {
	rv := make(map[string]any)
	return &SeederImpl{
		runtimeVars: RuntimeVars{
			vars: rv,
		},
		log: log,
	}
}

func (r *SeederImpl) WithClient(c Client) *SeederImpl {
	r.client = c
	return r
}

// WithAuth assigns auth options used by AuthMapRef within the actions
func (r *SeederImpl) WithAuth(a AuthMap) *SeederImpl {
	r.auth = NewAuth(a)
	return r
}

func (r *SeederImpl) SetRuntimeVar(key string, val any) {
	r.runtimeVars.mu.Lock()
	r.runtimeVars.vars[key] = val
	r.runtimeVars.mu.Unlock()
}

func (r *SeederImpl) RuntimeVars() map[string]any {
	rv := make(map[string]any)
	r.runtimeVars.mu.RLock()
	for k, v := range r.runtimeVars.vars {
		rv[k] = v
	}
	r.runtimeVars.mu.RUnlock()
	return rv
}

// +k8s:deepcopy-gen=true
// StrategyConfig defines top level Spec
type StrategyConfig struct {
	AuthConfig AuthMap `yaml:"auth" json:"auth"`
	Seeders    Seeders `yaml:"seed" json:"seed"`
}

// +k8s:deepcopy-gen=true
type Seeders map[string]Action

// NOTE: currently need to generate the below functions manually
// To allow for both integer or string in x-kubernetes unknown fields

// +k8s:deepcopy-gen=false
type KvMapVarsAny map[string]any

func (in *KvMapVarsAny) DeepCopyInto(out *KvMapVarsAny) {
	if in == nil {
		*out = nil
	} else {
		// *out = runtime.RawExtension(*in)
		*out = runtime.DeepCopyJSON(*in)
	}
}

func (in *KvMapVarsAny) DeepCopy() *KvMapVarsAny {
	if in == nil {
		return nil
	}
	out := new(KvMapVarsAny)
	in.DeepCopyInto(out)
	return out
}

// +k8s:deepcopy-gen=true
//
// Action defines the single action to make agains an endpoint
// and selecting a strategy
//
// Endpoint is the base url to make the requests against
// GetEndpointSuffix can be used to specify a direct ID or query params
// PostEndpointSuffix
type Action struct {
	name                 string             `yaml:"-" json:"-"`
	templatedPayload     string             `yaml:"-" json:"-"`
	foundId              string             `yaml:"-" json:"-"`
	header               *http.Header       `yaml:"-" json:"-"`
	Strategy             string             `yaml:"strategy" json:"strategy"`
	Order                *int               `yaml:"order,omitempty" json:"order,omitempty"`
	Endpoint             string             `yaml:"endpoint" json:"endpoint"`
	GetEndpointSuffix    *string            `yaml:"getEndpointSuffix,omitempty" json:"getEndpointSuffix,omitempty"`
	PostEndpointSuffix   *string            `yaml:"postEndpointSuffix,omitempty" json:"postEndpointSuffix,omitempty"`
	PatchEndpointSuffix  *string            `yaml:"patchEndpointSuffix,omitempty" json:"patchEndpointSuffix,omitempty"`
	PutEndpointSuffix    *string            `yaml:"putEndpointSuffix,omitempty" json:"putEndpointSuffix,omitempty"`
	DeleteEndpointSuffix *string            `yaml:"deleteEndpointSuffix,omitempty" json:"deleteEndpointSuffix,omitempty"`
	FindByJsonPathExpr   string             `yaml:"findByJsonPathExpr,omitempty" json:"findByJsonPathExpr,omitempty"`
	PayloadTemplate      string             `yaml:"payloadTemplate" json:"payloadTemplate"`
	PatchPayloadTemplate string             `yaml:"patchPayloadTemplate,omitempty" json:"patchPayloadTemplate,omitempty"`
	RuntimeVars          map[string]string  `yaml:"runtimeVars,omitempty" json:"runtimeVars,omitempty"`
	AuthMapRef           string             `yaml:"authMapRef" json:"authMapRef"`
	HttpHeaders          *map[string]string `yaml:"httpHeaders,omitempty" json:"httpHeaders,omitempty"`
	Variables            KvMapVarsAny       `yaml:"variables" json:"variables"`
}

// WithHeader allows the overwrite of default Accept and Content-Type headers
// both default to `application/json` and adding additional header params on
// per Action basis. NOTE: each rest call inside the action
// will inherit the same header
func (a *Action) WithHeader() *Action {
	suppliedHeader := a.HttpHeaders

	h := &http.Header{}
	// set default values
	h.Add("Accept", "application/json")
	h.Add("Content-Type", "application/json")
	// overwrite or add additional header attributes on a per action basis
	if suppliedHeader != nil {
		for k, v := range *suppliedHeader {
			h.Add(k, v)
		}
	}

	a.header = h
	return a
}

func (a *Action) WithName(name string) *Action {
	a.name = name
	return a
}

type Status int

const (
	StatusFatal Status = iota
	StatusRetryable
)

// do performs all the network calls -
// each request object is with Context
// re-use same context with auth implementation
func (r *SeederImpl) do(req *http.Request, action *Action) ([]byte, error) {
	r.log.Debug("starting request")
	r.log.Debugf("request: %+v", req)
	respBody := []byte{}
	req.Header = *action.header
	diag := &Diagnostic{HostPathMethod: fmt.Sprintf("Method => %s HostPath => %s%s Query => %s", req.Method, req.URL.Host, req.URL.Path, req.URL.RawQuery), Name: action.name, ProceedFallback: false, IsFatal: true}

	r.log.Debugf("restPayload diagnostic: %+v", diag)

	resp, err := r.client.Do(r.doAuth(req, action))
	if err != nil {
		r.log.Debugf("failed to make network call: %v", err)
		diag.WithStatus(999) // networkError
		diag.WithMessage(fmt.Sprintf("failed to make network call: %v", err.Error()))
		r.log.Debugf("diagnostic: %+v", diag)
		return nil, diag
	}
	defer resp.Body.Close()

	diag.WithStatus(resp.StatusCode)

	if resp.Body != nil {
		if respBody, err = io.ReadAll(resp.Body); err != nil {
			r.log.Debugf("failed to read body, closed or empty")
			diag.WithMessage("unable to read the body")
			return nil, diag
		}
	}
	// in case we need to follow redirects (shouldn't really for backend calls...)
	if resp.StatusCode > 299 {
		diag.WithMessage(string(respBody))
		diag.WithProceedFallback(true)
		diag.WithIsFatal(false)
		r.log.Debugf("resp status code: %d", resp.StatusCode)
		if resp.StatusCode >= 500 {
			r.log.Debugf("resp status code: %d, most likely indicates a service down. should not proceed to further strategies", resp.StatusCode)
			diag.WithProceedFallback(false)
		}
		return nil, diag
	}
	// every successful response should be passed through
	// setRunTimeVar to ensure we grab anything highlighted by users
	r.findRuntimeVars(respBody, action)
	return respBody, nil
}

func (r *SeederImpl) doAuth(req *http.Request, action *Action) *http.Request {
	enrichedReq := req
	am := *r.auth
	switch cam := am[action.AuthMapRef]; cam.authStrategy {
	case Basic:
		enrichedReq.SetBasicAuth(cam.basicAuth.username, cam.basicAuth.password)
	case OAuth:
		token, err := cam.oAuthConfig.Token(enrichedReq.Context())
		if err != nil {
			r.log.Errorf("failed to obtain token: %v", err)
		}
		enrichedReq.Header.Set("Authorization", fmt.Sprintf("%s %s", token.TokenType, token.AccessToken))
	case OAuthPassword:
		token, err := cam.passwordGrantConfig.oauthPassCredsConfig.PasswordCredentialsToken(enrichedReq.Context(), cam.passwordGrantConfig.resourceOwnerUser,
			cam.passwordGrantConfig.resourceOwnerPass)
		if err != nil {
			r.log.Errorf("failed to obtain token: %v", err)
		}
		enrichedReq.Header.Set("Authorization", fmt.Sprintf("%s %s", token.TokenType, token.AccessToken))
	case CustomToToken:
		token, err := cam.customToToken.Token(enrichedReq.Context(), r.client, r.log)
		if err != nil {
			r.log.Errorf("failed to obtain custom token: %v", err)
		}
		enrichedReq.Header.Set(token.HeaderKey, fmt.Sprintf("%s %s", token.TokenPrefix, token.TokenValue))
	case StaticToken:
		enrichedReq.Header.Set(cam.staticToken.headerKey, cam.staticToken.staticToken)
	}
	return enrichedReq
}

// get makes a network call on caller defined client.Do
// returns the body as byte array
func (r *SeederImpl) get(ctx context.Context, action *Action) ([]byte, error) {
	endpoint := action.Endpoint
	if action.GetEndpointSuffix != nil {
		endpoint = fmt.Sprintf("%s%s", endpoint, *action.GetEndpointSuffix)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)

	if err != nil {
		r.log.Debugf("failed to build request: %v", err)
	}

	return r.do(req, action)
}

func (r *SeederImpl) post(ctx context.Context, action *Action) error {
	endpoint := action.Endpoint
	if action.PostEndpointSuffix != nil {
		endpoint = fmt.Sprintf("%s%s", endpoint, *action.PostEndpointSuffix)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(action.templatedPayload))

	if err != nil {
		r.log.Error(err)
	}

	if _, err := r.do(req, action); err != nil {
		return err
	}
	return nil
}

func (r *SeederImpl) patch(ctx context.Context, action *Action) error {
	endpoint := action.Endpoint
	if action.PatchEndpointSuffix != nil {
		endpoint = fmt.Sprintf("%s%s", endpoint, *action.PatchEndpointSuffix)
	}
	if action.foundId != "" {
		endpoint = fmt.Sprintf("%s/%s", endpoint, action.foundId)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", endpoint, strings.NewReader(action.templatedPayload))

	if err != nil {
		r.log.Error(err)
	}

	if _, err := r.do(req, action); err != nil {
		return err
	}
	return nil
}

func (r *SeederImpl) put(ctx context.Context, action *Action) error {
	// create a local reference copy in each base call
	endpoint := action.Endpoint
	if action.PutEndpointSuffix != nil {
		endpoint = fmt.Sprintf("%s%s", endpoint, *action.PutEndpointSuffix)
	}
	if action.foundId != "" {
		endpoint = fmt.Sprintf("%s/%s", endpoint, action.foundId)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, strings.NewReader(action.templatedPayload))

	if err != nil {
		r.log.Error(err)
	}
	if _, err := r.do(req, action); err != nil {
		return err
	}

	return nil
}

// deleteMethod calls base do and returns error if any failures
func (r *SeederImpl) delete(ctx context.Context, action *Action) error {
	endpoint := action.Endpoint
	if action.DeleteEndpointSuffix != nil {
		endpoint = fmt.Sprintf("%s%s", endpoint, *action.DeleteEndpointSuffix)
	}
	if action.foundId != "" {
		endpoint = fmt.Sprintf("%s/%s", endpoint, action.foundId)
	}
	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)

	if err != nil {
		r.log.Error(err)
	}

	if _, err := r.do(req, action); err != nil {
		return err
	}
	return nil
}

func (r *SeederImpl) FindPathByExpression(resp []byte, pathExpression string) (string, error) {
	return findPathByExpression(resp, pathExpression, r.log)
}

// findPathByExpression lookup func using jsonpathexpression
func findPathByExpression(resp []byte, pathExpression string, log log.Loggeriface) (string, error) {
	unescStr := "" //string(resp)
	if pathExpression == "" {
		log.Info("no path expression provided returning empty")
		return "", nil
	}

	unescStr, err := strconv.Unquote(fmt.Sprintf("%v", string(resp)))
	if err != nil {
		log.Debug("using original string")
		unescStr = string(resp)
	}

	result, err := ajson.JSONPath([]byte(unescStr), pathExpression)
	if err != nil {
		log.Debug("failed to perform JSON path lookup - epxression failure")
		return "", err
	}

	for _, v := range result {
		switch v.Type() {
		// case int, int64, int32, int16, int8, float64, float32:
		case ajson.String:
			str, e := strconv.Unquote(fmt.Sprintf("%v", v))
			if e != nil {
				log.Debugf("unable to unquote value: %v returning as is", v)
				return fmt.Sprintf("%v", v), e
			}
			return str, nil
		case ajson.Numeric:
			return fmt.Sprintf("%v", v), nil
		default:
			return "", fmt.Errorf("cannot use type: %v in further processing - can only be a numeric or string value", v.Type())
		}
	}
	log.Infof("expression not yielded any results")
	return "", nil
}

// TemplatePayload parses input payload and replaces all $var ${var} with
// existing global env variable as well as injected from inside RestAction
// into the local context
func (r *SeederImpl) TemplatePayload(payload string, vars KvMapVarsAny) string {
	localVars := &KvMapVarsAny{}
	if vars == nil {
		vars = *localVars
	}

	// extend existing to allow for runtimeVars replacement
	for k, v := range r.RuntimeVars() {
		vars[k] = v
	}
	tmpl, err := templatePayload(payload, vars)
	if err != nil {
		r.log.Errorf("unable to parse template: %v", err)
	}
	return tmpl
}

func templatePayload(payload string, vars KvMapVarsAny) (string, error) {
	for k, v := range vars {
		os.Setenv(k, fmt.Sprintf("%v", v))
	}

	// // double escape any $ or ${ conforming to this regex
	// // so that it is not picked up by envsubst substitution
	// envSubstEscaped := regexp.MustCompile(`[^\$]\$[a-zA-Z\{]`).ReplaceAllStringFunc(payload, func(entry string) string {
	// 	return strings.Replace(entry, "$", "$$", -1)
	// })
	return envsubst.StringRestrictedNoDigit(payload, false, false, true)
}

// findRuntimeVars checks the response for any runtime variables
// supplied by caller in the initial template
func (r *SeederImpl) findRuntimeVars(createUpdateResponse []byte, action *Action) {
	if action.RuntimeVars == nil {
		return
	}
	// checks runtimeVars in the action
	// assigns any values found to the SeederImpl
	// makes it available for all subsequent actions
	for k, v := range action.RuntimeVars {
		found, err := r.FindPathByExpression(createUpdateResponse, v)
		if err != nil {
			r.log.Errorf("error finding pathexpr in runtime var")
			r.log.Debugf("failed on: %v, with expr: %v", k, v)
			r.log.Debugf("continuing...")
		}
		if found != "" {
			r.SetRuntimeVar(k, found)
		}
	}
}
