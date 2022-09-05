package strategyrestseeder

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	log "github.com/dnitsch/simplelog"
	"github.com/dnitsch/strategyrestseeder/internal/testutils"
	yaml "gopkg.in/yaml.v3"
)

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

var testYaml = []byte(`
auth:
  type: basicAuth
  username: user
  password: pass
seed:
  find-put-post-not-found-id:
    endpoint: https://postman-echo.com
    strategy: FIND/PUT/POST
    getEndpointSuffix: /get?json=provided
    postEndpointSuffix: /post
    payloadTemplate: |
      {
      	"value": "$foo"
      }
    variables:
      foo: bar
    runtimeVars:
      someId: "$.array[?(@.name=='fubar')].id"
  find-put-post-found-id:
    endpoint: https://postman-echo.com
    strategy: FIND/PUT/POST
    getEndpointSuffix: /get?json=provided
    postEndpointSuffix: /post
    putEndpointSuffix: /put
    findByJsonPathExpr: "$.array[?(@.name=='fubar')].id"
    payloadTemplate: |
      {
      	"value": "$foo"
      }
    variables:
      foo: bar
    runtimeVars:
      someId: "$.array[?(@.name=='fubar')].id"
`)

func Test_Execute(t *testing.T) {
	tests := []struct {
		name string
		srs  func(t *testing.T) *StrategyRestSeeder
	}{
		{
			name: "successful find put post rest calls",
			srs: func(t *testing.T) *StrategyRestSeeder {
				strategy := StrategyConfig{}
				srs := New()

				// yInpt, _ := os.ReadFile(".ignore-test.yaml")

				_ = yaml.Unmarshal(testYaml, &strategy)

				mc := &MockClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {

						if req.Method == "GET" && !strings.Contains(req.URL.Path, "/get") {
							t.Errorf(testutils.TestPhrase, "https://postman-echo.com/get", req.RequestURI)
						}
						if req.Method == "POST" && !strings.Contains(req.URL.Path, "/post") {
							t.Errorf(testutils.TestPhrase, "https://postman-echo.com/post", req.RequestURI)
						}
						if req.Method == "PUT" && !strings.Contains(req.URL.Path, "/put") {
							t.Errorf(testutils.TestPhrase, "https://postman-echo.com/put", req.RequestURI)
						}
						return &http.Response{
							Body:       io.NopCloser(strings.NewReader(`{"array":[{"id":3,"name":"fubar","a":"b","c":"d"},{"id":32,"name":"fubar2","a":"f","c":"h"},{"id":42,"name":"fubar42","a":"i","c":"j"}]}`)),
							StatusCode: 200,
						}, nil
					},
				}
				srs.WithActions(strategy.Seeders).WithLogger(os.Stdout, log.DebugLvl).WithRestClient(mc)

				return srs
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if e := tt.srs(t).Execute(); e != nil {
				t.Error(e)
			}

		})
	}
}
