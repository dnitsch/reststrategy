package seeder_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
		"OAuth Client Creds GET/PUT/POST rest success": {
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
						AuthMapRef:         "oauth2-test",
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
	}
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

func TestExecuteFindPutPost(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) seeder.AuthMap
		expect     func(url string) string
		seeders    func(url string) seeder.Seeders
	}{
		"OAuth PasswordCredentials FIND/PUT/POST success": {
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
					"find-put-post-found": {
						Strategy:           string(seeder.FIND_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/all"),
						PostEndpointSuffix: seeder.String("/post"),
						PutEndpointSuffix:  seeder.String("/put"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
					"find-put-post-not-found": {
						Strategy:           string(seeder.FIND_PUT_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/not-found"),
						PostEndpointSuffix: seeder.String("/post/new"),
						PutEndpointSuffix:  seeder.String("/put/not-found"),
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						PayloadTemplate:    `{"value": "$foo"}`,
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
				mux.HandleFunc("/token", OAuthPasswordHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
	}
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
					"get-post-found": {
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
					"get-post-not-found": {
						Strategy:           string(seeder.GET_POST),
						Order:              seeder.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  seeder.String("/get/not-found"),
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
		"OAuth Client Creds PUT/POST rest success": {
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
				mux.HandleFunc("/token", testutils.TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
	}
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

func TestExecutePutPost(t *testing.T) {

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
				mux.HandleFunc("/token", testutils.TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
	}
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

func TestExecuteFindPatchPost(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) seeder.AuthMap
		expect     func(url string) string
		seeders    func(url string) seeder.Seeders
	}{
		"OAuth Client Creds FIND/PATCH/POST rest success": {
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
						AuthMapRef:           "oauth2-test",
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
				mux.HandleFunc("/token", testutils.TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
	}
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
				mux.HandleFunc("/token", testutils.TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
	}
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
				})
				mux.HandleFunc("/token", testutils.OAuthPasswordHandleFunc(t))
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
				mux.HandleFunc("/token", testutils.TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
	}
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
				mux.HandleFunc("/token", testutils.OAuthPasswordHandleFunc(t))
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
				mux.HandleFunc("/token", testutils.TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) string {
				return ""
			},
		},
	}
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