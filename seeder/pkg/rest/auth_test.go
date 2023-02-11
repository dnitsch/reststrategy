package rest_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/dnitsch/reststrategy/seeder/internal/testutils"
	"github.com/dnitsch/reststrategy/seeder/pkg/rest"
	log "github.com/dnitsch/simplelog"
)

type mockClient func(req *http.Request) (*http.Response, error)

func (m mockClient) Do(req *http.Request) (*http.Response, error) {
	return m(req)
}
func TestCustomToken(t *testing.T) {
	ttests := map[string]struct {
		customAuth func(t *testing.T) *rest.CustomFlowAuth
		client     func(t *testing.T) rest.Client
		expect     rest.CustomTokenResponse
	}{
		"success with default": {
			func(t *testing.T) *rest.CustomFlowAuth {
				return rest.NewCustomFlowAuth().WithAuthUrl("http://foo.bar").WithAuthMap(rest.KvMapVarsAny{"email": "bar@qux.boo", "pass": "barrr"})
			},
			func(t *testing.T) rest.Client {
				return mockClient(func(req *http.Request) (*http.Response, error) {
					resp := &http.Response{Body: io.NopCloser(strings.NewReader(`{"access_token":"8s0ghews87ghv78gh8ergh8dfgfdg"}`))}
					return resp, nil
				})
			},
			rest.CustomTokenResponse{
				TokenPrefix: "Bearer",
				TokenValue:  "8s0ghews87ghv78gh8ergh8dfgfdg",
				HeaderKey:   "Authorization",
			},
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.customAuth(t).Token(context.TODO(), tt.client(t), log.New(&bytes.Buffer{}, log.DebugLvl))
			if err != nil {
				t.Errorf(testutils.TestPhrase, got, tt.expect)
			}
			if got.TokenValue != tt.expect.TokenValue {
				t.Errorf(testutils.TestPhraseWithContext, "custom token not returned properly", got.TokenValue, tt.expect.TokenValue)
			}
		})
	}
}

func TestClientCredentials(t *testing.T) {
	ttests := map[string]struct {
		am rest.AuthMap
	}{
		"clientcreds": {
			am: map[string]rest.AuthConfig{
				"client": {
					AuthStrategy: rest.OAuth,
					Username:     "foo",
					Password:     "bar",
					OAuth: &rest.ConfigOAuth{
						ServerUrl:               "http://test.bar",
						Scopes:                  []string{"profile"},
						EndpointParams:          map[string][]string{"foo": {"bar", "bax"}},
						OAuthSendParamsInHeader: false,
						ResourceOwnerUser:       nil,
						ResourceOwnerPassword:   nil,
					},
				},
			},
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			got := rest.NewAuth(tt.am)
			if got == nil {
				t.Errorf(testutils.TestPhraseWithContext, "auth map", "not nil", "<nil>")
			}
		})
	}
}
