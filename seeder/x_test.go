// +go:build !ignore
package seeder_test

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

const (
	TestPhrase            string = "got: %v want: %v\n"
	TestPhraseWithContext string = "%s got: %v want: %v\n"
)

func TokenHandleFunc(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
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

func OAuthPasswordHandleFunc(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		wantGrantType := "password"
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
