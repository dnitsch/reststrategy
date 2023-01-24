package seeder

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dnitsch/reststrategy/seeder/internal/testutils"
	"github.com/dnitsch/reststrategy/seeder/pkg/rest"
	log "github.com/dnitsch/simplelog"
)

func Test_Execute(t *testing.T) {

	ts := httptest.NewServer(testutils.TestMuxServer(t))
	logW := &bytes.Buffer{}

	logger := log.New(logW, log.DebugLvl)

	tests := []struct {
		name             string
		srs              func(t *testing.T) *StrategyRestSeeder
		authConfig       *rest.AuthMap
		expectErrorCount int
		seeders          rest.Seeders
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
			seeders: rest.Seeders{
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
				return New(&logger).WithRestClient(&http.Client{})
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
			seeders: rest.Seeders{
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
				return New(&logger).WithRestClient(&http.Client{})
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
			seeders: rest.Seeders{
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
				return New(&logger).WithRestClient(&http.Client{})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_srs := tt.srs(t)
			_srs.WithActions(tt.seeders).WithAuth(tt.authConfig)
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
