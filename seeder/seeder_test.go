package seeder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/dnitsch/reststrategy/seeder/pkg/rest"
	log "github.com/dnitsch/simplelog"
)

func token(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		wantGrantType := "client_credentials"
		expiry := time.Now()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll(r.Body) == %v, %v, want _, <nil>", body, err)
		}
		// if err := r.Body.Close(); err != nil {
		// 	t.Errorf("r.Body.Close() == %v, want <nil>", err)
		// }
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Errorf("url.ParseQuery(%q) == %v, %v, want _, <nil>", body, values, err)
		}
		gotGrantType := values.Get("grant_type")
		if gotGrantType != wantGrantType {
			t.Errorf("grant_type == %q, want %q", gotGrantType, wantGrantType)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"access_token":"90d64460d14870c08c81352a05dedd3465940a7c","token_type":"bearer","expiry": "%s"}`, expiry.Add(1*time.Hour).String())))
	}
}

func tokenInHeader(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		wantGrantType := "client_credentials"
		expiry := time.Now()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll(r.Body) == %v, %v, want _, <nil>", body, err)
		}
		// if err := r.Body.Close(); err != nil {
		// 	t.Errorf("r.Body.Close() == %v, want <nil>", err)
		// }
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Errorf("url.ParseQuery(%q) == %v, %v, want _, <nil>", body, values, err)
		}
		gotGrantType := values.Get("grant_type")
		if gotGrantType != wantGrantType {
			t.Errorf("grant_type == %q, want %q", gotGrantType, wantGrantType)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"access_token":"90d64460d14870c08c81352a05dedd3465940a7c","token_type":"bearer","expiry": "%s"}`, expiry.Add(1*time.Hour).String())))
	}
}

func getHandle(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		switch values, _ := url.ParseQuery(r.URL.RawQuery); values.Get("simulate_resp") {
		case "single_valid":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id":3,"name":"fubar","a":"b","c":"d"}`))
		case "all_valid":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"array":[{"id":3,"name":"fubar","a":"b","c":"d"},{"id":32,"name":"fubar2","a":"f","c":"h"},{"id":42,"name":"fubar42","a":"i","c":"j"}]}`))
		case "single_empty":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{}`))
		case "all_empty":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"array":[]}`))
		case "bad_request":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{}`))
		case "error":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{}`))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{}`))
		}
	}
}

func putHandle(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("io.ReadAll(r.Body) == %v, %v, want _, <nil>", body, err)
		}
		if err := r.Body.Close(); err != nil {
			t.Errorf("r.Body.Close() == %v, want <nil>", err)
		}
		if unreplacedFound := bytes.Contains(body, []byte(`$`)); unreplacedFound {
			t.Errorf("body: %q,should not contain any unreplaced $", string(body))
		}
		switch values, _ := url.ParseQuery(r.URL.RawQuery); values.Get("simulate_resp") {
		case "not_found":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{}`))
		case "bad_request":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{}`))
		case "error":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{}`))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{}`))
		}
	}
}

func postHandle(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("io.ReadAll(r.Body) == %v, %v, want _, <nil>", body, err)
		}
		if err := r.Body.Close(); err != nil {
			t.Errorf("r.Body.Close() == %v, want <nil>", err)
		}
		if unreplacedFound := bytes.Contains(body, []byte(`$`)); unreplacedFound {
			t.Errorf("body: %q,should not contain any unreplaced $", string(body))
		}
		switch values, _ := url.ParseQuery(r.URL.RawQuery); values.Get("simulate_resp") {
		case "not_found":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{}`))
		case "bad_request":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{}`))
		case "error":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{}`))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{}`))
		}
	}
}

func deletetHandle(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("io.ReadAll(r.Body) == %v, %v, want _, <nil>", body, err)
		}
		if err := r.Body.Close(); err != nil {
			t.Errorf("r.Body.Close() == %v, want <nil>", err)
		}
		if unreplacedFound := bytes.Contains(body, []byte(`$`)); unreplacedFound {
			t.Errorf("body: %q,should not contain any unreplaced $", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
		w.WriteHeader(http.StatusAccepted)
	}
}

func Test_Execute(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/token", token(t))
	mux.HandleFunc("/token/header", tokenInHeader(t))
	mux.HandleFunc("/get/", getHandle(t))
	mux.HandleFunc("/put", putHandle(t))
	mux.HandleFunc("/put/", putHandle(t))
	mux.HandleFunc("/post", postHandle(t))
	mux.HandleFunc("/delete", deletetHandle(t))
	mux.HandleFunc("/delete/", deletetHandle(t))

	ts := httptest.NewServer(mux)

	tests := []struct {
		name             string
		srs              func(t *testing.T) *StrategyRestSeeder
		authConfig       *rest.AuthMap
		expectErrorCount int
		seeders          Seeders
	}{
		{
			name: "OAuth find put post rest calls",
			authConfig: &rest.AuthMap{
				"oauth2-test": {
					AuthStrategy: rest.OAuth,
					Username:     "randClientIdOrUsernameForBasicAuth",
					Password:     "randClientSecretOrPassExpr",
					OAuth: &rest.ConfigOAuth{
						OAuthSendParamsInHeader: false,
						ServerUrl:               fmt.Sprintf("%s/token", ts.URL),
						Scopes:                  []string{"foo", "bar"},
						EndpointParams:          map[string][]string{"params": {"baz", "boom"}},
					},
				},
			},
			seeders: Seeders{
				"get-put-post-found": {
					Strategy:           string(GET_PUT_POST),
					Order:              rest.Int(0),
					Endpoint:           ts.URL,
					GetEndpointSuffix:  rest.String("/get/single?simulate_resp=single_valid"),
					PostEndpointSuffix: rest.String("/post"),
					PutEndpointSuffix:  rest.String("/put"),
					PayloadTemplate:    `{"value": "$foo"}`,
					Variables:          map[string]any{"foo": "bar"},
					RuntimeVars:        &map[string]string{"someId": "$.array[?(@.name=='fubar')].id"},
					AuthMapRef:         "oauth2-test",
				},
				"find-put-post-found": {
					Strategy:           string(FIND_PUT_POST),
					Order:              rest.Int(0),
					Endpoint:           ts.URL,
					GetEndpointSuffix:  rest.String("/get/all?simulate_resp=all_valid"),
					PostEndpointSuffix: rest.String("/post"),
					PutEndpointSuffix:  rest.String("/put"),
					FindByJsonPathExpr: "$.array[?(@.name=='fubar')].id",
					PayloadTemplate:    `{"value": "$foo"}`,
					Variables:          map[string]any{"foo": "bar"},
					AuthMapRef:         "oauth2-test",
					RuntimeVars:        &map[string]string{"someId": "$.array[?(@.name=='fubar')].id"},
				},
				"find-put-post-empty-not-found": {
					Strategy:           string(FIND_PUT_POST),
					Order:              rest.Int(0),
					Endpoint:           ts.URL,
					GetEndpointSuffix:  rest.String("/get/all/empty?simulate_resp=all_empty"),
					PostEndpointSuffix: rest.String("/post"),
					PutEndpointSuffix:  rest.String("/put"),
					FindByJsonPathExpr: "$.array[?(@.name=='fubar')].id",
					PayloadTemplate:    `{"value": "$foo"}`,
					Variables:          map[string]any{"foo": "bar"},
					AuthMapRef:         "oauth2-test",
					RuntimeVars:        &map[string]string{"someId": "$.array[?(@.name=='fubar')].id"},
				},
				"find-put-post-bad-request": {
					Strategy:           string(FIND_PUT_POST),
					Order:              rest.Int(0),
					Endpoint:           ts.URL,
					GetEndpointSuffix:  rest.String("/get/all/empty?simulate_resp=bad_request"),
					PostEndpointSuffix: rest.String("/post"),
					PutEndpointSuffix:  rest.String("/put"),
					FindByJsonPathExpr: "$.array[?(@.name=='fubar')].id",
					PayloadTemplate:    `{"value": "$foo"}`,
					Variables:          map[string]any{"foo": "bar"},
					AuthMapRef:         "oauth2-test",
					RuntimeVars:        &map[string]string{"someId": "$.array[?(@.name=='fubar')].id"},
				},
			},
			expectErrorCount: 1,
			srs: func(t *testing.T) *StrategyRestSeeder {
				return New().WithRestClient(&http.Client{})
			},
		},
		{
			name: "BasicAuth calls PUT",
			authConfig: &rest.AuthMap{
				"basic-test": {
					AuthStrategy: rest.Basic,
					Username:     "randClientIdOrUsernameForBasicAuth",
					Password:     "randClientSecretOrPassExpr",
				},
			},
			seeders: Seeders{
				"put": {
					Strategy:           string(PUT),
					Order:              rest.Int(0),
					Endpoint:           ts.URL,
					GetEndpointSuffix:  nil,
					PostEndpointSuffix: nil,
					PutEndpointSuffix:  rest.String("/put/id-1234"),
					PayloadTemplate:    `{"value": "$foo"}`,
					Variables:          map[string]any{"foo": "bar"},
					AuthMapRef:         "basic-test",
				},
				"put-post-found": {
					Strategy:           string(PUT_POST),
					Order:              rest.Int(1),
					Endpoint:           ts.URL,
					GetEndpointSuffix:  nil,
					PostEndpointSuffix: rest.String("/post"),
					PutEndpointSuffix:  rest.String("/put/id-1234"),
					PayloadTemplate:    `{"id": "id-1234","value": "$foo"}`,
					Variables:          map[string]any{"foo": "bar"},
					AuthMapRef:         "basic-test",
				},
				"put-post-empty-not-found": {
					Strategy:           string(PUT_POST),
					Order:              rest.Int(2),
					Endpoint:           ts.URL,
					PostEndpointSuffix: rest.String("/post"),
					PutEndpointSuffix:  rest.String("/put/id-1234?simulate_resp=not_found"),
					PayloadTemplate:    `{"id": "id-1234", "value": "$foo"}`,
					Variables:          map[string]any{"foo": "bar"},
					AuthMapRef:         "basic-test",
				},
				"put-bad-request": {
					Strategy:           string(PUT_POST),
					Order:              rest.Int(2),
					Endpoint:           ts.URL,
					PostEndpointSuffix: rest.String("/post"),
					PutEndpointSuffix:  rest.String("/put/id-1234?simulate_resp=bad_request"),
					PayloadTemplate:    `{"id": "id-1234", "value": "$foo"}`,
					Variables:          map[string]any{"foo": "bar"},
					AuthMapRef:         "basic-test",
				},
				"put-server-error": {
					Strategy:           string(PUT_POST),
					Order:              rest.Int(2),
					Endpoint:           ts.URL,
					PostEndpointSuffix: rest.String("/post"),
					PutEndpointSuffix:  rest.String("/put/id-1234?simulate_resp=error"),
					PayloadTemplate:    `{"value": "$foo"}`,
					Variables:          map[string]any{"foo": "bar"},
					AuthMapRef:         "basic-test",
				},
			},
			expectErrorCount: 1,
			srs: func(t *testing.T) *StrategyRestSeeder {
				return New().WithRestClient(&http.Client{})
			},
		},
		{
			name: "Mix BasicAuth and OAuth calls find-delete-post",
			authConfig: &rest.AuthMap{
				"oauth2-test": {
					AuthStrategy: rest.OAuth,
					Username:     "randClientIdOrUsernameForBasicAuth",
					Password:     "randClientSecretOrPassExpr",
					OAuth: &rest.ConfigOAuth{
						OAuthSendParamsInHeader: false,
						ServerUrl:               fmt.Sprintf("%s/token", ts.URL),
					},
				},
				"basic-test": {
					AuthStrategy: rest.Basic,
					Username:     "randClientIdOrUsernameForBasicAuth",
					Password:     "randClientSecretOrPassExpr",
				},
			},
			seeders: Seeders{
				"not-found-do-post": {
					Strategy:             string(FIND_DELETE_POST),
					Order:                rest.Int(0),
					Endpoint:             ts.URL,
					GetEndpointSuffix:    rest.String("/get/all/empty?simulate_resp=all_empty"),
					PostEndpointSuffix:   rest.String("/post"),
					DeleteEndpointSuffix: rest.String("/delete"),
					FindByJsonPathExpr:   "$.array[?(@.name=='fubar')].id",
					PayloadTemplate:      `{"value": "$foo"}`,
					Variables:            map[string]any{"foo": "bar"},
					AuthMapRef:           "oauth2-test",
					HttpHeaders:          &map[string]string{"X-Foo": "bar"},
				},
				"found-do-delete-then-post": {
					Strategy:             string(FIND_DELETE_POST),
					Order:                rest.Int(0),
					Endpoint:             ts.URL,
					GetEndpointSuffix:    rest.String("/get/all/empty?simulate_resp=all_valid"),
					PostEndpointSuffix:   rest.String("/post"),
					DeleteEndpointSuffix: rest.String("/delete"),
					FindByJsonPathExpr:   "$.array[?(@.name=='fubar')].id",
					PayloadTemplate:      `{"value": "$foo"}`,
					Variables:            map[string]any{"foo": "bar"},
					AuthMapRef:           "basic-test",
				},
				"found-do-delete-then-post-error": {
					Strategy:             string(FIND_DELETE_POST),
					Order:                rest.Int(0),
					Endpoint:             ts.URL,
					GetEndpointSuffix:    rest.String("/get/all/empty?simulate_resp=all_valid"),
					PostEndpointSuffix:   rest.String("/post?simulate_resp=bad_request"),
					DeleteEndpointSuffix: rest.String("/delete"),
					FindByJsonPathExpr:   "$.array[?(@.name=='fubar')].id",
					PayloadTemplate:      `{"value": "$foo"}`,
					Variables:            map[string]any{"foo": "bar"},
					AuthMapRef:           "oauth2-test",
				},
			},
			expectErrorCount: 1,
			srs: func(t *testing.T) *StrategyRestSeeder {
				return New().WithRestClient(&http.Client{})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logW := &bytes.Buffer{}
			_srs := tt.srs(t)

			_srs.WithActions(tt.seeders).WithAuth(tt.authConfig).WithLogger(logW, log.DebugLvl)
			// ctx, cancel := context.WithCancel(context.TODO())
			// defer cancel()
			e := _srs.Execute(context.Background())
			if tt.expectErrorCount != len(e) {
				t.Error("error counts do not match")
			}
			if e != nil {
				if tt.expectErrorCount > 0 {
					if tt.expectErrorCount != len(e) {
						t.Errorf("expected error count: %d, got: %d", tt.expectErrorCount, len(e))
					}
				} else {
					t.Error(e)
				}
			}
		})
	}
}

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	// just in case you want default correct return value
	return &http.Response{}, nil
}

// var testYaml = []byte(`
// auth:
//   type: OAuthClientCredentials
//   username: randClientIdOrUsernameForBasicAuth
//   password: randClientSecret
//   oauth:
//     serverUrl: http://localhost:8080/token
//     scopes:
//       - https://www.some-api-provider.com/scopes-example1
//     endpointParams:
//       foo:
//         - bar
//         - baz
//   httpHeaders:
//     X-Foo: bar
// seed:
//   find-put-post-not-found-id:
//     endpoint: https://postman-echo.com
//     strategy: FIND/PUT/POST
//     getEndpointSuffix: /get/empty?json=emtpy
//     postEndpointSuffix: /post
//     findByJsonPathExpr: "$.array[?(@.name=='fubar')].id"
//     payloadTemplate: |
//       { "value": "$foo" }
//     variables:
//       foo: bar
//     runtimeVars:
//       someId: "$.array[?(@.name=='fubar')].id"
//   find-put-post-found-id:
//     endpoint: https://postman-echo.com
//     strategy: FIND/PUT/POST
//     getEndpointSuffix: /get/valid?json=provided
//     postEndpointSuffix: /post
//     putEndpointSuffix: /put
//     findByJsonPathExpr: "$.array[?(@.name=='fubar')].id"
//     payloadTemplate: |
//       {
//       	"value": "$foo"
//       }
//     variables:
//       foo: bar
//     runtimeVars:
//       someId: "$.array[?(@.name=='fubar')].id"
//   get-put-post-found-id:
//     endpoint: https://postman-echo.com
//     strategy: GET/PUT/POST
//     getEndpointSuffix: /get/single-valid?json=provided
//     postEndpointSuffix: /post
//     putEndpointSuffix: /put
//     payloadTemplate: |
//       { "value": "$foo"}
//     variables:
//       foo: bar
//   get-put-post-not-found-id:
//     endpoint: https://postman-echo.com
//     strategy: GET/PUT/POST
//     getEndpointSuffix: /get/single-empty?json=provided
//     postEndpointSuffix: /post
//     putEndpointSuffix: /put
//     payloadTemplate: |
//       { "value": "$foo" }
//     variables:
//       foo: bar
// `)
