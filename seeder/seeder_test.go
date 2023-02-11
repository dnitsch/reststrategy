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
	"github.com/dnitsch/reststrategy/seeder/internal/testutils"
	"github.com/dnitsch/reststrategy/seeder/pkg/rest"
	log "github.com/dnitsch/simplelog"
)

func TestExecute(t *testing.T) {

	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := map[string]struct {
		handler    func(t *testing.T) http.Handler
		authConfig func(url string) rest.AuthMap
		expect     func(url string) error
		seeders    func(url string) rest.Seeders
	}{
		"OAuth Client Creds GET/PUT/POST rest success": {
			authConfig: func(url string) rest.AuthMap {
				return rest.AuthMap{
					"oauth2-test": {
						AuthStrategy: rest.OAuth,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &rest.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
						},
					},
				}
			},
			seeders: func(url string) rest.Seeders {
				return rest.Seeders{
					"get-put-post-found": {
						Strategy:           string(seeder.GET_PUT_POST),
						Order:              rest.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  rest.String("/get/1234"),
						PostEndpointSuffix: rest.String("/post"),
						PutEndpointSuffix:  rest.String("/put/1234"),
						PayloadTemplate:    `{"value": "$foo"}`,
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-test",
					},
					"get-put-post-not-found": {
						Strategy:           string(seeder.GET_PUT_POST),
						Order:              rest.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  rest.String("/get/not-found"),
						PostEndpointSuffix: rest.String("/post/new"),
						PutEndpointSuffix:  rest.String("/put/not-found"),
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
				mux.HandleFunc("/token", testutils.TokenHandleFunc(t))
				return mux
			},
			expect: func(url string) error {
				return nil
				// 	//fmt.Errorf(`status: 400
				// name: find-put-post-bad-request
				// message: {}
				// hostPathMethod: Method => GET HostPath => %s/get/all/empty Query => simulate_resp=bad_request
				// isRetryAble: true`, strings.TrimPrefix(url, "http://"))
			},
		},
		"OAuth PAsswordCredentials FIND/PUT/POST success": {
			authConfig: func(url string) rest.AuthMap {
				return rest.AuthMap{
					"oauth2-passwd": {
						AuthStrategy: rest.OAuthPassword,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
						OAuth: &rest.ConfigOAuth{
							OAuthSendParamsInHeader: false,
							ServerUrl:               fmt.Sprintf("%s/token", url),
							Scopes:                  []string{"foo", "bar"},
							EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
							ResourceOwnerUser:       rest.String("bob"),
							ResourceOwnerPassword:   rest.String("barfooqux"),
						},
					},
				}
			},
			seeders: func(url string) rest.Seeders {
				return rest.Seeders{
					"find-put-post-found": {
						Strategy:           string(seeder.FIND_PUT_POST),
						Order:              rest.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  rest.String("/get/all"),
						PostEndpointSuffix: rest.String("/post"),
						PutEndpointSuffix:  rest.String("/put"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "oauth2-passwd",
					},
					"find-put-post-not-found": {
						Strategy:           string(seeder.FIND_PUT_POST),
						Order:              rest.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  rest.String("/get/not-found"),
						PostEndpointSuffix: rest.String("/post/new"),
						PutEndpointSuffix:  rest.String("/put/not-found"),
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
				mux.HandleFunc("/token", testutils.OAuthPasswordHandleFunc(t))
				return mux
			},
			expect: func(url string) error {
				return nil
			},
		},
		"BasicAuth Get/Post success": {
			authConfig: func(url string) rest.AuthMap {
				return rest.AuthMap{
					"basic": {
						AuthStrategy: rest.Basic,
						Username:     "randClientIdOrUsernameForBasicAuth",
						Password:     "randClientSecretOrPassExpr",
					},
				}
			},
			seeders: func(url string) rest.Seeders {
				return rest.Seeders{
					"get-post-found": {
						Strategy:           string(seeder.GET_POST),
						Order:              rest.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  rest.String("/get/1234"),
						PostEndpointSuffix: rest.String("/post"),
						PayloadTemplate:    `{"value": "$foo"}`,
						FindByJsonPathExpr: "$.[?(@.name=='fubar')].id",
						Variables:          map[string]any{"foo": "bar"},
						AuthMapRef:         "basic",
					},
					"get-post-not-found": {
						Strategy:           string(seeder.GET_POST),
						Order:              rest.Int(0),
						Endpoint:           url,
						GetEndpointSuffix:  rest.String("/get/not-found"),
						PostEndpointSuffix: rest.String("/post/new"),
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
			expect: func(url string) error {
				return nil
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
				if err.Error() != tt.expect(ts.URL).Error() {
					t.Errorf("expected different error got: %v\n\nwant: %v", err.Error(), tt.expect(ts.URL).Error())
				}
			}
		})
	}
}
