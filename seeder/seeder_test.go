package seeder_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/reststrategy/seeder"
	log "github.com/dnitsch/simplelog"
)

func TestExecuteGetPutPost(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) seeder.AuthMap
		expect     func(url string) string
		seeders    func(url string) seeder.Seeders
	}{
		"OAuth/CustomToken Client Creds GET/PUT/POST rest success": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
					"custom-to-token-test": {
						AuthStrategy: seeder.CustomToToken,
						CustomToken: &seeder.CustomToken{
							CustomAuthMap: map[string]any{"email": "some@one.com", "password": "p4ssword", "grant_type": "client_credentials"},
							SendInHeader:  true,
							AuthUrl:       fmt.Sprintf("%s/customToken", url),
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"get-put-post-found": {
						Strategy:           string(seeder.GET_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/1234"),
						PostEndpointSuffix: seeder.String("/post"),
						PutEndpointSuffix:  seeder.String("/put/1234"),
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-test",
					},
					"get-put-post-not-found": {
						Strategy:           string(seeder.GET_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/not-found"),
						PostEndpointSuffix: seeder.String("/post/new"),
						PutEndpointSuffix:  seeder.String("/put/not-found"),
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "custom-to-token-test",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(404)
				})
				mux.HandleFunc("/get/1234", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name": "fubar", "id":"1234"}`))
				})
				mux.HandleFunc("/put/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")

					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/put/not-found", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("put should never be called but was called with: %v", string(b))
				})
				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				mux.HandleFunc("/customToken", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
				// 	//fmt.Errorf(`status: 400
				// name: find-put-post-bad-request
				// message: {}
				// hostPathMethod: Method => GET HostPath => %s/get/all/empty Query => simulate_resp=bad_request
				// isRetryAble: true`, strings.TrimPrefix(url, "http://"))
			},
		},
		"CustomToken GET/PUT/POST GET 500 error": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"custom-to-token-test": {
						AuthStrategy: seeder.CustomToToken,
						CustomToken: &seeder.CustomToken{
							CustomAuthMap: map[string]any{"email": "some@one.com", "password": "p4ssword", "grant_type": "client_credentials"},
							SendInHeader:  true,
							AuthUrl:       fmt.Sprintf("%s/customToken", url),
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"get-put-post-get-error": {
						Strategy:           string(seeder.GET_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PutEndpointSuffix:  seeder.String("/put"),
						PostEndpointSuffix: seeder.String("/post"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "custom-to-token-test",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(500)
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/put/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "newBar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "newBar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")

					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/customToken", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "status: 500"
			},
		},
		"CustomToken GET/PUT/POST GET empty response": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"custom-to-token-test": {
						AuthStrategy: seeder.CustomToToken,
						CustomToken: &seeder.CustomToken{
							CustomAuthMap: map[string]any{"email": "some@one.com", "password": "p4ssword", "grant_type": "client_credentials"},
							SendInHeader:  true,
							AuthUrl:       fmt.Sprintf("%s/customToken", url),
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"get-put-post-get-error": {
						Strategy:           string(seeder.GET_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PutEndpointSuffix:  seeder.String("/put"),
						PostEndpointSuffix: seeder.String("/post/new"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "custom-to-token-test",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(``))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/customToken", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "unexpected end of file"
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srs := seeder.New(&logger).WithRestClient(&http.Client{})

			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()

			srs.WithActions(tt.seeders(ts.URL)).WithAuth(tt.authConfig(ts.URL))

			err := srs.Execute(context.Background())

			if err != nil {
				if !strings.HasPrefix(err.Error(), tt.expect(ts.URL)) {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL))
				}
			}
		})
	}
}

func TestExecuteFindPutPost(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) seeder.AuthMap
		expect     func(url string) string
		seeders    func(url string) seeder.Seeders
	}{
		"CustomToken ClientCredentials FIND/PUT/POST success": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"custom-to-token-test": {
						AuthStrategy: seeder.CustomToToken,
						CustomToken: &seeder.CustomToken{
							CustomAuthMap: map[string]any{"email": "some@one.com", "password": "p4ssword", "grant_type": "client_credentials"},
							SendInHeader:  true,
							AuthUrl:       fmt.Sprintf("%s/customToken", url),
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"find-put-post-found-runtime-vars-found": {
						Strategy:           string(seeder.FIND_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PostEndpointSuffix: seeder.String("/post"),
						PutEndpointSuffix:  seeder.String("/put"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						RuntimeVars:        map[string]string{"runFoo": "$.[?(@.name=='fubar')].id"},
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "custom-to-token-test",
					},
					"find-put-post-not-found": {
						Strategy:           string(seeder.FIND_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/not-found"),
						PostEndpointSuffix: seeder.String("/post/new"),
						PutEndpointSuffix:  seeder.String("/put/not-found"),
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "custom-to-token-test",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[]`))
				})
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[{"name":"fubar","id":"1234"}]`))
				})
				mux.HandleFunc("/put/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")

					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/put/not-found", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("put should never be called but was called with: %v", string(b))
				})
				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/customToken", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
		"OAuth2 FIND/PUT/POST GET 500 error": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"find-put-post-get-error": {
						Strategy:           string(seeder.FIND_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PostEndpointSuffix: seeder.String("/put"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
					"find-put-post-get-error-blankJsonPathExpr": {
						Strategy:           string(seeder.FIND_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PostEndpointSuffix: seeder.String("/put"),
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(500)
				})
				mux.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "status: 500"
			},
		},
		"OAuth2 FIND/PUT/POST GET empty response": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"find-post-get-error": {
						Strategy:           string(seeder.FIND_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PutEndpointSuffix:  seeder.String("/put/1234"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
					"find-post-get-error-blankJsonPathExpr": {
						Strategy:           string(seeder.FIND_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PostEndpointSuffix: seeder.String("/put/1234"),
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(``))
				})
				mux.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/put/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "unexpected end of file"
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srs := seeder.New(&logger).WithRestClient(&http.Client{})

			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()

			srs.WithActions(tt.seeders(ts.URL)).WithAuth(tt.authConfig(ts.URL))

			err := srs.Execute(context.Background())

			if err != nil {
				if !strings.HasPrefix(err.Error(), tt.expect(ts.URL)) {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL))
				}
			}
		})
	}
}

func TestExecuteCustomTokenError(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) seeder.AuthMap
		expect     func(url string) string
		seeders    func(url string) seeder.Seeders
	}{
		"CustomToken error": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"custom-to-token-error": {
						AuthStrategy: seeder.CustomToToken,
						CustomToken: &seeder.CustomToken{
							CustomAuthMap: map[string]any{"email": 1, "password": "p4ssword", "grant_type": "password"},
							SendInHeader:  true,
							AuthUrl:       fmt.Sprintf("%s/customToken", url),
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"custom-to-token-error-find-put-post-found-runtime-vars-found": {
						Strategy:           string(seeder.FIND_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           "url",
						GetEndpointSuffix:  seeder.String("/get/all"),
						PostEndpointSuffix: seeder.String("/post"),
						PutEndpointSuffix:  seeder.String("/put"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						RuntimeVars:        map[string]string{"runFoo": "$.[?(@.name=='fubar')].id"},
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "custom-to-token-test",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[]`))
				})
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[{"name":"fubar","id":"1234"}]`))
				})
				mux.HandleFunc("/put/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")

					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/put/not-found", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("put should never be called but was called with: %v", string(b))
				})
				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/customToken", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srs := seeder.New(&logger).WithRestClient(&http.Client{})

			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()

			srs.WithActions(tt.seeders(ts.URL)).WithAuth(tt.authConfig(ts.URL))

			err := srs.Execute(context.Background())

			if err != nil {
				if !strings.HasPrefix(err.Error(), tt.expect(ts.URL)) {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL))
				}
			}
		})
	}
}

func TestExecuteGetPost(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) seeder.AuthMap
		expect     func(url string) string
		seeders    func(url string) seeder.Seeders
	}{
		"BasicAuth Get/Post success": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"basic": {
						AuthStrategy: seeder.Basic,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"get-post-found-set": {
						Strategy:           string(seeder.GET_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/1234"),
						PostEndpointSuffix: seeder.String("/post"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						RuntimeVars:        map[string]string{"runFoo": "runBar"},
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "basic",
					},
					"get-post-not-found-not-set-runtime-vars": {
						Strategy:           string(seeder.GET_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/not-found"),
						PostEndpointSuffix: seeder.String("/post/new"),
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						RuntimeVars:        map[string]string{"runFoo": "runBar"},
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "basic",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/not-found", func(w http.ResponseWriter, r *http.Request) {
					if v, ok := r.Header["Authorization"]; !ok {
						t.Errorf("basic auth header not set got: %v, wanted: Basic base64(username:password)", v)
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[]`))
				})
				mux.HandleFunc("/get/1234", func(w http.ResponseWriter, r *http.Request) {
					if v, ok := r.Header["Authorization"]; !ok {
						t.Errorf("basic auth header not set got: %v, wanted: Basic base64(username:password)", v)
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
		"BasicAuth Get/Post 500 error": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"basic": {
						AuthStrategy: seeder.Basic,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"get-post-500-error": {
						Strategy:           string(seeder.GET_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/1234"),
						PostEndpointSuffix: seeder.String("/post"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "basic",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/1234", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(500)
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				return mux
			},
			expect: func(url string) string {
				return "status: 500"
			},
		},
		"BasicAuth Get/Post Empty Get Response": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"basic": {
						AuthStrategy: seeder.Basic,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"empty-get-post": {
						Strategy:           string(seeder.GET_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/1234"),
						PostEndpointSuffix: seeder.String("/post/new"),
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "basic",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/1234", func(w http.ResponseWriter, r *http.Request) {
					if v, ok := r.Header["Authorization"]; !ok {
						t.Errorf("basic auth header not set got: %v, wanted: Basic base64(username:password)", v)
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(``))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
		"OAuth Client Creds Get/Post Empty Get Response": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"empty-get-post": {
						Strategy:           string(seeder.GET_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/1234"),
						PostEndpointSuffix: seeder.String("/post/new"),
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-test",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/1234", func(w http.ResponseWriter, r *http.Request) {
					if v, ok := r.Header["Authorization"]; !ok {
						t.Errorf("basic auth header not set got: %v, wanted: Basic base64(username:password)", v)
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(``))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
		"OAuth Client Creds Get Error creating new request": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"empty-get-post": {
						Strategy:           string(seeder.GET_POST),
						Order:              seeder.Int(0),
						Endpoint:           string([]byte{0x7f}),
						GetEndpointSuffix:  seeder.String(""),
						PostEndpointSuffix: seeder.String("/post/new"),
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-test",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/1234", func(w http.ResponseWriter, r *http.Request) {
					if v, ok := r.Header["Authorization"]; !ok {
						t.Errorf("basic auth header not set got: %v, wanted: Basic base64(username:password)", v)
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(``))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return `parse "\x7f": net/url: invalid control character in URL`
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srs := seeder.New(&logger).WithRestClient(&http.Client{})

			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()

			srs.WithActions(tt.seeders(ts.URL)).WithAuth(tt.authConfig(ts.URL))

			err := srs.Execute(context.Background())

			if err != nil {
				if !strings.HasPrefix(err.Error(), tt.expect(ts.URL)) {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL))
				}
			}
		})
	}
}

func TestExecutePutPost(t *testing.T) {
	flag.Set("test.timeout", "2m30s")

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) seeder.AuthMap
		expect     func(url string) string
		seeders    func(url string) seeder.Seeders
	}{
		"OAuth Client Creds PUT/POST rest success": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: true,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
					"no-auth": {
						AuthStrategy: seeder.NoAuth,
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"put-post-found": {
						Strategy:           string(seeder.PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						PostEndpointSuffix: seeder.String("/post"),
						PutEndpointSuffix:  seeder.String("/put/1234"),
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-test",
					},
					"put-post-not-found": {
						Strategy:           string(seeder.PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						PostEndpointSuffix: seeder.String("/post/new"),
						PutEndpointSuffix:  seeder.String("/put/not-found"),
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-test",
					},
					"put-post-error-network": {
						Strategy:           string(seeder.PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           "http://unknown.domain.foo:34567",
						PostEndpointSuffix: seeder.String("/post/new"),
						PutEndpointSuffix:  seeder.String("/put/not-found"),
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "no-auth",
					},
					"put-post-tls-error": {
						Strategy:           string(seeder.PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           "https://unknown.domain.foo:34567",
						PostEndpointSuffix: seeder.String("/post/new"),
						PutEndpointSuffix:  seeder.String("/put/not-found"),
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "no-auth",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/put/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/put/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(404)
				})
				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "status: 999"
			},
		},
		"BasicAuth PUT/POST 500 error": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"basic": {
						AuthStrategy: seeder.Basic,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"put-post-500-error": {
						Strategy:           string(seeder.PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/1234"),
						PostEndpointSuffix: seeder.String("/post"),
						PutEndpointSuffix:  seeder.String("/put"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "basic",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/put/1234", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(500)
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(500)
				})
				return mux
			},
			expect: func(url string) string {
				return "status: 500"
			},
		},
		"OAuth2 PUT/POST GET 500 error": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"put-post-get-error": {
						Strategy:           string(seeder.PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						PutEndpointSuffix:  seeder.String("/put/1234"),
						PostEndpointSuffix: seeder.String("/post"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/put/1234", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(500)
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "status: 500"
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srs := seeder.New(&logger).WithRestClient(&http.Client{})

			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()

			srs.WithActions(tt.seeders(ts.URL)).WithAuth(tt.authConfig(ts.URL))

			want := tt.expect(ts.URL)

			if err := srs.Execute(context.Background()); err != nil {
				if want == "" {
					t.Fatalf("unexpected error occurred: %s", err.Error())
				}
				if !strings.HasPrefix(err.Error(), want) {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL))
				}
			}
		})
	}
}

func TestExecuteFindPatchPost(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) seeder.AuthMap
		expect     func(url string) string
		seeders    func(url string) seeder.Seeders
	}{
		"StaticToken FIND/PATCH/POST rest success": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"static-token-test": {
						AuthStrategy: seeder.StaticToken,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"patch-post-found": {
						Strategy:             string(seeder.FIND_PATCH_POST),
						Order:                seeder.Int(0),
						Endpoint:             url,
						GetEndpointSuffix:    seeder.String("/get/all"),
						PatchEndpointSuffix:  seeder.String("/patch"),
						PostEndpointSuffix:   seeder.String("/post"),
						PayloadTemplate:      `{"value": "$foo"}`,
						PatchPayloadTemplate: `{"value": "$newFoo"}`,
						FindByJsonPathExpr:   "$.[?(@.name=='fubar')].id",
						Variables:            map[string]any{"foo": "bar", "newFoo": "newBar"},
						AuthMapRef:           "static-token-test",
					},
					"patch-post-not-found": {
						Strategy:             string(seeder.FIND_PATCH_POST),
						Order:                seeder.Int(0),
						Endpoint:             url,
						GetEndpointSuffix:    seeder.String("/get/not-found"),
						PatchEndpointSuffix:  seeder.String("/patch/not-found"),
						PostEndpointSuffix:   seeder.String("/post/new"),
						PayloadTemplate:      `{"value": "$foo"}`,
						PatchPayloadTemplate: `{"value": "$newFoo"}`,
						FindByJsonPathExpr:   "$.[?(@.name=='fubar')].id",
						Variables:            map[string]any{"foo": "bar", "newFoo": "newBar"},
						AuthMapRef:           "static-token-test",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[]`))
				})
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[{"name":"fubar","id":"1234"}]`))
				})
				mux.HandleFunc("/patch/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "newBar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "newBar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")

					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/patch/not-found", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("patch should never be called but was called with: %v", string(b))
				})
				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
		"OAuth2 FIND/PATCH/POST GET 500 error": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"find-patch-post-get-error": {
						Strategy:            string(seeder.FIND_PATCH_POST),
						Order:               seeder.Int(0),
						Endpoint:            url,
						GetEndpointSuffix:   seeder.String("/get/all"),
						PatchEndpointSuffix: seeder.String("/patch"),
						PostEndpointSuffix:  seeder.String("/post"),
						PayloadTemplate:     `{"value": "$foo"}`,
						FindByJsonPathExpr:  "$.[?(@.name=='fubar')].id",
						Variables:           map[string]any{"foo": "bar"},
						AuthMapRef:          "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(500)
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/patch/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "newBar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "newBar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")

					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "status: 500"
			},
		},
		"OAuth2 FIND/PATCH/POST GET empty response": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"find-patch-post-get-error": {
						Strategy:            string(seeder.FIND_PATCH_POST),
						Order:               seeder.Int(0),
						Endpoint:            url,
						GetEndpointSuffix:   seeder.String("/get/all"),
						PatchEndpointSuffix: seeder.String("/patch"),
						PostEndpointSuffix:  seeder.String("/post"),
						PayloadTemplate:     `{"value": "$foo"}`,
						FindByJsonPathExpr:  "$.[?(@.name=='fubar')].id",
						Variables:           map[string]any{"foo": "bar"},
						AuthMapRef:          "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(``))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "unexpected end of file"
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srs := seeder.New(&logger).WithRestClient(&http.Client{})

			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()

			srs.WithActions(tt.seeders(ts.URL)).WithAuth(tt.authConfig(ts.URL))

			err := srs.Execute(context.Background())

			if err != nil {
				if !strings.HasPrefix(err.Error(), tt.expect(ts.URL)) {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL))
				}
			}
		})
	}
}

func TestExecuteFindDeletePost(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) seeder.AuthMap
		expect     func(url string) string
		seeders    func(url string) seeder.Seeders
	}{
		"OAuth Client Creds FIND/DELETE/POST rest success": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"delete-post-found": {
						Strategy:             string(seeder.FIND_DELETE_POST),
						Order:                seeder.Int(0),
						Endpoint:             url,
						GetEndpointSuffix:    seeder.String("/get/all"),
						DeleteEndpointSuffix: seeder.String("/delete"),
						PostEndpointSuffix:   seeder.String("/post/1234"),
						PayloadTemplate:      `{"value": "$foo"}`,
						FindByJsonPathExpr:   "$.[?(@.name=='fubar')].id",
						Variables:            map[string]any{"foo": "bar"},
						AuthMapRef:           "oauth2-test",
					},
					"delete-post-not-found": {
						Strategy:             string(seeder.FIND_DELETE_POST),
						Order:                seeder.Int(0),
						Endpoint:             url,
						GetEndpointSuffix:    seeder.String("/get/not-found"),
						DeleteEndpointSuffix: seeder.String("/delete/not-found"),
						PostEndpointSuffix:   seeder.String("/post/1234"),
						PayloadTemplate:      `{"value": "$foo"}`,
						FindByJsonPathExpr:   "$.[?(@.name=='fubar')].id",
						Variables:            map[string]any{"foo": "bar"},
						AuthMapRef:           "oauth2-test",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[]`))
				})
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[{"name":"fubar","id":"1234"}]`))
				})
				mux.HandleFunc("/delete/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if len(string(b)) > 0 {
						t.Errorf(`got: %v Expected body to be empty`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/delete/not-found", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("delete should never be called but was called with: %v", string(b))
				})
				mux.HandleFunc("/post/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
		"OAuth Client Creds FIND/DELETE/POST delete error": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"find-delete-post-delete-error": {
						Strategy:             string(seeder.FIND_DELETE_POST),
						Order:                seeder.Int(0),
						Endpoint:             url,
						GetEndpointSuffix:    seeder.String("/get/all"),
						DeleteEndpointSuffix: seeder.String("/delete"),
						PayloadTemplate:      `{"value": "$foo"}`,
						FindByJsonPathExpr:   "$.[?(@.name=='fubar')].id",
						Variables:            map[string]any{"foo": "bar"},
						AuthMapRef:           "oauth2-test",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[{"name":"fubar","id":"1234"}]`))
				})
				mux.HandleFunc("/delete/1234", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(500)
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/delete/not-found", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("delete should never be called but was called with: %v", string(b))
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "status: 500"
			},
		},
		"OAuth2 FIND/DELETE/POST GET 500 error": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"find-delete-post-get-error": {
						Strategy:             string(seeder.FIND_DELETE_POST),
						Order:                seeder.Int(0),
						Endpoint:             url,
						GetEndpointSuffix:    seeder.String("/get/all"),
						DeleteEndpointSuffix: seeder.String("/delete"),
						PostEndpointSuffix:   seeder.String("/post"),
						PayloadTemplate:      `{"value": "$foo"}`,
						FindByJsonPathExpr:   "$.[?(@.name=='fubar')].id",
						Variables:            map[string]any{"foo": "bar"},
						AuthMapRef:           "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(500)
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/delete/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "newBar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "newBar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")

					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "status: 500"
			},
		},
		"OAuth2 FIND/DELETE/POST GET 302 Redirect": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"find-delete-post-get-error": {
						Strategy:             string(seeder.FIND_DELETE_POST),
						Order:                seeder.Int(0),
						Endpoint:             url,
						GetEndpointSuffix:    seeder.String("/get/all"),
						DeleteEndpointSuffix: seeder.String("/delete"),
						PostEndpointSuffix:   seeder.String("/post"),
						PayloadTemplate:      `{"value": "$foo"}`,
						FindByJsonPathExpr:   "$.[?(@.name=='fubar')].id",
						Variables:            map[string]any{"foo": "bar"},
						AuthMapRef:           "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(302)
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/delete/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "newBar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "newBar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "unexpected end of file"
			},
		},
		"OAuth2 FIND/DELETE/POST GET empty response": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"find-delete-post-get-error": {
						Strategy:             string(seeder.FIND_DELETE_POST),
						Order:                seeder.Int(0),
						Endpoint:             url,
						GetEndpointSuffix:    seeder.String("/get/all"),
						DeleteEndpointSuffix: seeder.String("/delete"),
						PostEndpointSuffix:   seeder.String("/post"),
						PayloadTemplate:      `{"value": "$foo"}`,
						FindByJsonPathExpr:   "$.[?(@.name=='fubar')].id",
						Variables:            map[string]any{"foo": "bar"},
						AuthMapRef:           "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(``))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "unexpected end of file"
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srs := seeder.New(&logger).WithRestClient(&http.Client{})

			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()

			srs.WithActions(tt.seeders(ts.URL)).WithAuth(tt.authConfig(ts.URL))

			err := srs.Execute(context.Background())

			if err != nil {
				if !strings.HasPrefix(err.Error(), tt.expect(ts.URL)) {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL))
				}
			}
		})
	}
}

func TestExecutePut(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) seeder.AuthMap
		expect     func(url string) string
		seeders    func(url string) seeder.Seeders
	}{
		"OAuth PasswordCredentials PUT success": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-passwd": {
						AuthStrategy: seeder.OAuthPassword,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
							ResourceOwnerUser:       seeder.String("bob"),
							ResourceOwnerPassword:   seeder.String("barfooqux"),
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"put-found": {
						Strategy:           string(seeder.PUT),
						Order:              seeder.Int(0),
						Endpoint:           url,
						PutEndpointSuffix:  seeder.String("/put/1234"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/put/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")

					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/put/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(404)
					w.WriteHeader(404)
				})
				mux.HandleFunc("/token", OAuthPasswordHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
		"OAuth2 PUT success": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"put-found": {
						Strategy:           string(seeder.PUT),
						Order:              seeder.Int(0),
						Endpoint:           url,
						PutEndpointSuffix:  seeder.String("/put/1234"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-test",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/put/1234", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/put/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(404)
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srs := seeder.New(&logger).WithRestClient(&http.Client{})

			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()

			srs.WithActions(tt.seeders(ts.URL)).WithAuth(tt.authConfig(ts.URL))

			err := srs.Execute(context.Background())

			if err != nil {
				if err.Error() != tt.expect(ts.URL) {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL))
				}
			}
		})
	}
}

func TestExecuteFindPost(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) seeder.AuthMap
		expect     func(url string) string
		seeders    func(url string) seeder.Seeders
	}{
		"OAuth PasswordCredentials FIND/POST success": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-passwd": {
						AuthStrategy: seeder.OAuthPassword,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
							ResourceOwnerUser:       seeder.String("bob"),
							ResourceOwnerPassword:   seeder.String("barfooqux"),
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"post-found": {
						Strategy:           string(seeder.FIND_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PostEndpointSuffix: seeder.String("/post"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
					"post-not-found": {
						Strategy:           string(seeder.FIND_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/not-found"),
						PostEndpointSuffix: seeder.String("/post/new"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[]`))
				})
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[{"name":"fubar","id":"1234"}]`))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})

				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}

					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/token", OAuthPasswordHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
		"OAuth2 FIND/POST success": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"post-found": {
						Strategy:           string(seeder.FIND_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PostEndpointSuffix: seeder.String("/post"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
					"post-not-found": {
						Strategy:           string(seeder.FIND_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/not-found"),
						PostEndpointSuffix: seeder.String("/post/new"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[]`))
				})
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[{"name":"fubar","id":"1234"}]`))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
		"OAuth2 FIND/POST GET 500 error": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"find-post-get-error": {
						Strategy:           string(seeder.FIND_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PostEndpointSuffix: seeder.String("/post"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(500)
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "status: 500"
			},
		},
		"OAuth2 FIND/POST GET empty response": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-test": {
						AuthStrategy: seeder.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"find-post-get-error": {
						Strategy:           string(seeder.FIND_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PostEndpointSuffix: seeder.String("/post"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(``))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})
				mux.HandleFunc("/token", TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return "unexpected end of file"
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srs := seeder.New(&logger).WithRestClient(&http.Client{})

			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()

			srs.WithActions(tt.seeders(ts.URL)).WithAuth(tt.authConfig(ts.URL))

			err := srs.Execute(context.Background())

			if err != nil {
				if !strings.HasPrefix(err.Error(), tt.expect(ts.URL)) {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL))
				}
			}
		})
	}
}

type CMMock func(input string, config generator.GenVarsConfig) (string, error)

func (c CMMock) RetrieveWithInputReplaced(input string, config generator.GenVarsConfig) (string, error) {
	return c(input, config)
}

func TestExecuteWithConfigManager(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler       func(t *testing.T) http.Handler
		configmanager func(t *testing.T) seeder.CMRetrieve
		authConfig    func(url string) seeder.AuthMap
		expect        func(url string) string
		seeders       func(url string) seeder.Seeders
	}{
		"No Auth FIND/POST replace in template": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{}
			},
			configmanager: func(t *testing.T) seeder.CMRetrieve {
				return CMMock(func(input string, config generator.GenVarsConfig) (string, error) {
					return `{"name":"post-not-found","strategy":"FIND/POST","order":0,"endpoint":"http://127.0.0.1:62324","getEndpointSuffix":"/get/not-found","postEndpointSuffix":"/post/new","findByJsonPathExpr":"$.[?(@.name=='fubar')].id","payloadTemplate":"{\"value\": \"bar\"}","authMapRef":"basic","variables":null}`, nil
				})
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"post-not-found": {
						Strategy:           string(seeder.FIND_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/not-found"),
						PostEndpointSuffix: seeder.String("/post/new"),
						PayloadTemplate:    `{"value": "AWSPARAMSTR:///bar/foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						AuthMapRef:         "basic",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[]`))
				})
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[{"name":"fubar","id":"1234"}]`))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})

				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
		"No Auth FIND/POST replace in vars": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{}
			},
			configmanager: func(t *testing.T) seeder.CMRetrieve {
				return CMMock(func(input string, config generator.GenVarsConfig) (string, error) {

					return `{"name": "post-not-found-replaced-from-vars","strategy":"FIND/POST","order":0,"endpoint":"http://127.0.0.1:62324","getEndpointSuffix":"/get/not-found","postEndpointSuffix":"/post/new","findByJsonPathExpr":"$.[?(@.name=='fubar')].id","payloadTemplate":"{\"value\": \"$foo\"}","authMapRef":"basic","variables":{"foo":"bar"}}`, nil
				})
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"post-not-found-replaced-from-vars": {
						Strategy:           string(seeder.FIND_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/not-found"),
						PostEndpointSuffix: seeder.String("/post/new"),
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "AWSPARAMSTR:///bar/foo"},
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						AuthMapRef:         "basic",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[]`))
				})
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[{"name":"fubar","id":"1234"}]`))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})

				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					if string(b) != `{"value": "bar"}` {
						t.Errorf(`got: %v expected body to match the templated payload: {"value": "bar"}`, string(b))
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
		"Basic  Auth FIND/POST replace in vars": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"basic": seeder.AuthConfig{
						Username:     "foo",
						Password:     "AWSPARAMSTR:///bar/foo",
						AuthStrategy: seeder.Basic,
					},
				}
			},
			configmanager: func(t *testing.T) seeder.CMRetrieve {
				return CMMock(func(input string, config generator.GenVarsConfig) (string, error) {
					return `{"basic": {"type":"BasicAuth","username":"foo","password":"bar"}}`, nil
				})
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {

			srs := seeder.New(&logger).WithRestClient(&http.Client{})

			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()

			srs.WithActions(tt.seeders(ts.URL)).WithAuth(tt.authConfig(ts.URL)).
				WithConfigManager(tt.configmanager(t)).WithConfigManagerOptions(generator.NewConfig().WithTokenSeparator("://"))

			err := srs.Execute(context.Background())

			if err != nil {
				if tt.expect(ts.URL) == "" {
					t.Errorf("did not expect an error to occur saw: %v ", err.Error())
				}
				if !strings.HasPrefix(err.Error(), tt.expect(ts.URL)) {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL))
				}
			}
		})
	}
}


func TestExecuteWithConfigManagerError(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler       func(t *testing.T) http.Handler
		configmanager func(t *testing.T) seeder.CMRetrieve
		authConfig    func(url string) seeder.AuthMap
		expect        func(url string) string
		seeders       func(url string) seeder.Seeders
	}{
		"Basic  Auth FIND/POST configmanager error response": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"basic": seeder.AuthConfig{
						Username:     "foo",
						Password:     "AWSPARAMSTR:///bar/foo",
						AuthStrategy: seeder.Basic,
					},
				}
			},
			configmanager: func(t *testing.T) seeder.CMRetrieve {
				return CMMock(func(input string, config generator.GenVarsConfig) (string, error) {
					return `{"basic": {"type":"BasicAuth","username":"foo","password":"bar"}}`, fmt.Errorf("Error occurred in configmanager")
				})
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				return mux
			},
			expect: func(url string) string {
				return "Error while replacing secrets placeholders in authmap - Error occurred in configmanager"
			},
		},
		"No Auth FIND/POST configmanager error": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{}
			},
			configmanager: func(t *testing.T) seeder.CMRetrieve {
				return CMMock(func(input string, config generator.GenVarsConfig) (string, error) {
					return `{"name":"post-not-found","strategy":"FIND/POST","order":0,"endpoint":"http://127.0.0.1:62324","getEndpointSuffix":"/get/not-found","postEndpointSuffix":"/post/new","findByJsonPathExpr":"$.[?(@.name=='fubar')].id","payloadTemplate":"{\"value\": \"bar\"}","authMapRef":"basic","variables":null}`, fmt.Errorf("Error occurred in configmanager")
				})
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"post-not-found": {
						Strategy:           string(seeder.FIND_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/not-found"),
						PostEndpointSuffix: seeder.String("/post/new"),
						PayloadTemplate:    `{"value": "AWSPARAMSTR:///bar/foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						AuthMapRef:         "basic",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get/not-found", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[]`))
				})
				mux.HandleFunc("/get/all", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`[{"name":"fubar","id":"1234"}]`))
				})
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					b, _ := io.ReadAll(r.Body)
					t.Errorf("post should never be called but was called with: %v", string(b))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
				})

				mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"name":"fubar","id":"1234"}`))
				})
				return mux
			},
			expect: func(url string) string {
				return "Error while replacing secrets placeholders in actions - Error occurred in configmanager"
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srs := seeder.New(&logger).WithRestClient(&http.Client{})

			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()

			srs.WithActions(tt.seeders(ts.URL)).WithAuth(tt.authConfig(ts.URL)).
				WithConfigManager(tt.configmanager(t)).WithConfigManagerOptions(generator.NewConfig().WithTokenSeparator("://"))

			err := srs.Execute(context.Background())

			if err != nil {
				if tt.expect(ts.URL) == "" {
					t.Errorf("did not expect an error to occur saw: %v ", err.Error())
				}
				if !strings.HasPrefix(err.Error(), tt.expect(ts.URL)) {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL))
				}
			}
		})
	}
}

func TestExecuteUnknownStrategy(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) seeder.AuthMap
		expect     func(url string) string
		seeders    func(url string) seeder.Seeders
	}{
		"OAuth PasswordCredentials Unknown Strategy": {
			authConfig: func(url string) seeder.AuthMap {
				return seeder.AuthMap{
					"oauth2-passwd": {
						AuthStrategy: seeder.OAuthPassword,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &seeder.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
							ResourceOwnerUser:       seeder.String("bob"),
							ResourceOwnerPassword:   seeder.String("barfooqux"),
						},
					},
				}
			},
			seeders: func(url string) seeder.Seeders {
				return seeder.Seeders{
					"post-found": {
						Strategy:           string("not-a-strategy"),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
				}
			},
			handler: func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/token", OAuthPasswordHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
	}
	t.Parallel()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srs := seeder.New(&logger).WithRestClient(&http.Client{})

			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()

			srs.WithActions(tt.seeders(ts.URL)).WithAuth(tt.authConfig(ts.URL))

			err := srs.Execute(context.Background())

			if err != nil {
				if err.Error() != tt.expect(ts.URL) {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL))
				}
			}
		})
	}
}
