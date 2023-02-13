package seeder_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/dnitsch/reststrategy/seeder"
	log "github.com/dnitsch/simplelog"
)

type mockClient func(req *http.Request) (*http.Response, error)

func (m mockClient) Do(req *http.Request) (*http.Response, error) {
	return m(req)
}
func TestCustomToken(t *testing.T) {
	ttests := map[string]struct {
		customAuth func(t *testing.T) *seeder.CustomFlowAuth
		client     func(t *testing.T) seeder.Client
		expect     seeder.CustomTokenResponse
	}{
		"success with default": {
			func(t *testing.T) *seeder.CustomFlowAuth {
				return seeder.NewCustomFlowAuth().WithAuthUrl("http://foo.bar").WithAuthMap(seeder.KvMapVarsAny{"email": "bar@qux.boo", "pass": "barrr"})
			},
			func(t *testing.T) seeder.Client {
				return mockClient(func(req *http.Request) (*http.Response, error) {
					resp := &http.Response{Body: io.NopCloser(strings.NewReader(`{"access_token":"8s0ghews87ghv78gh8ergh8dfgfdg"}`))}
					return resp, nil
				})
			},
			seeder.CustomTokenResponse{
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
				t.Errorf(TestPhrase, got, tt.expect)
			}
			if got.TokenValue != tt.expect.TokenValue {
				t.Errorf(TestPhraseWithContext, "custom token not returned properly", got.TokenValue, tt.expect.TokenValue)
			}
		})
	}
}

func TestClientCredentials(t *testing.T) {
	ttests := map[string]struct {
		am seeder.AuthMap
	}{
		"clientcreds": {
			am: map[string]seeder.AuthConfig{
				"client": {
					AuthStrategy: seeder.OAuth,
					Username:     "foo",
					Password:     "bar",
					OAuth: &seeder.ConfigOAuth{
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
			got := seeder.NewAuth(tt.am)
			if got == nil {
				t.Errorf(TestPhraseWithContext, "auth map", "not nil", "<nil>")
			}
		})
	}
}
