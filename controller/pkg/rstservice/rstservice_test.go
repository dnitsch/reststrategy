package rstservice_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dnitsch/reststrategy/apis/reststrategy/v1alpha1"
	"github.com/dnitsch/reststrategy/controller/pkg/rstservice"
	"github.com/dnitsch/reststrategy/seeder"
	log "github.com/dnitsch/simplelog"
)

const (
	TestPhrase         string = "got: %v want: %v"
	TestPhraseWContext string = "failed %s => got: %v want: %v"
)

func TestExecute(t *testing.T) {
	ttests := map[string]struct {
		spec    func(t *testing.T, url string) *v1alpha1.StrategySpec
		handler func(t *testing.T) http.Handler
	}{
		"succcess": {
			func(t *testing.T, url string) *v1alpha1.StrategySpec {
				return &v1alpha1.StrategySpec{
					AuthConfig: []v1alpha1.AuthConfig{
						{
							"basic",
							seeder.AuthConfig{
								AuthStrategy: seeder.Basic,
								Username:     "foo",
								Password:     "bar",
							},
						},
					},
					Seeders: []v1alpha1.SeederConfig{
						{
							"get-post-test",
							seeder.Action{
								Strategy:           "GET/POST",
								Endpoint:           url,
								GetEndpointSuffix:  seeder.String("/get"),
								PostEndpointSuffix: seeder.String("/post"),
								PayloadTemplate:    `{"foo":"bar"}`,
								AuthMapRef:         "basic",
							},
						},
					},
				}
			},
			func(t *testing.T) http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Write([]byte(`{"foo": "bar"}`))
				})
				return mux
			},
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler(t))
			defer ts.Close()
			rst := rstservice.New(log.New(&bytes.Buffer{}, log.ErrorLvl), &http.Client{})
			if err := rst.Execute(*tt.spec(t, ts.URL)); err != nil {
				t.Errorf(TestPhraseWContext, "execute failed with input=(%q)", err.Error(), "<nil>")
			}
		})
	}
}
