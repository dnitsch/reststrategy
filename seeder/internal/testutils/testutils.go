package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

const (
	TestPhrase string = "Want: %v\nGot: %v"
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
		case "echo":
			w.Header().Set("Content-Type", "application/json")
			args, err := json.Marshal(values)
			if err != nil {
				t.Error("failed to parse echo: %w", err.Error())
			}
			w.Write([]byte(fmt.Sprintf(`{"args":%v}`, args)))
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

func TestMuxServer(t *testing.T) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", token(t))
	mux.HandleFunc("/token/header", tokenInHeader(t))
	mux.HandleFunc("/get/", getHandle(t))
	mux.HandleFunc("/put", putHandle(t))
	mux.HandleFunc("/put/", putHandle(t))
	mux.HandleFunc("/post", postHandle(t))
	mux.HandleFunc("/delete", deletetHandle(t))
	mux.HandleFunc("/delete/", deletetHandle(t))
	mux.HandleFunc("/staticToken", getStaticTokenHandle(t))
	return mux
}

func getStaticTokenHandle(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if _, ok := r.Header["Token"]; !ok {
			t.Errorf("expected token to be set in header, got <nil>")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))

	}
}
