package rest

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dnitsch/reststrategy/seeder/internal/testutils"

	log "github.com/dnitsch/simplelog"
	"github.com/spyzhov/ajson"
)

func Test_getSeeder(t *testing.T) {
	ts := httptest.NewServer(testutils.TestMuxServer(t))
	tests := []struct {
		name   string
		action *Action
		rimpl  *SeederImpl
		auth   *AuthMap
		client Client
		expect string
	}{
		{
			name:   "getRestFunc_without_header",
			auth:   &AuthMap{"foo": {AuthStrategy: Basic, Username: "foo", Password: "bar"}},
			client: &http.Client{},
			rimpl:  &SeederImpl{runtimeVars: runtimeVars{}},
			action: &Action{
				PayloadTemplate:    "{}",
				Strategy:           "GET/POST",
				Endpoint:           ts.URL,
				GetEndpointSuffix:  String("/get/all/empty?simulate_resp=echo&id=32"),
				FindByJsonPathExpr: "$.args.id",
				AuthMapRef:         "foo",
				HttpHeaders:        nil,
			},
			expect: "32",
		},
		{
			name:   "getRestFunc_with_header",
			auth:   &AuthMap{"foo": {AuthStrategy: Basic, Username: "foo", Password: "bar"}},
			client: &http.Client{},
			rimpl:  &SeederImpl{runtimeVars: runtimeVars{}},
			action: &Action{
				PayloadTemplate:    "{}",
				Strategy:           "GET/POST",
				Endpoint:           ts.URL,
				GetEndpointSuffix:  String("/get/all/empty?simulate_resp=echo&id=32"),
				FindByJsonPathExpr: "$.args.id",
				AuthMapRef:         "foo",
				HttpHeaders:        &map[string]string{"foo": "bar"},
			},
			expect: "32",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0777)
			l := log.New(file, log.DebugLvl)
			tt.rimpl.log = l
			a := tt.action.WithHeader().WithName(tt.name)
			tt.rimpl.WithAuth(tt.auth).WithClient(tt.client)
			err := tt.rimpl.GetPost(context.TODO(), a)
			if err != nil {
				t.Errorf("failed %s", err)
			}
		})
	}
}

func Test_findByPathExpression(t *testing.T) {
	tests := []struct {
		name           string
		payload        []byte
		pathExpression string
		expect         string
		expErr         *string
	}{
		{
			name:           "single depth escaped",
			payload:        []byte(`"{\"args\":{\"id\":\"32\"},\"headers\":{\"x-forwarded-proto\":\"https\",\"x-forwarded-port\":\"443\",\"host\":\"postman-echo.com\",\"x-amzn-trace-id\":\"Root=1-63106cc4-6e8b66b055d278e5613db058\",\"content-length\":\"2\",\"accept\":\"application/json\",\"content-type\":\"application/json\",\"accept-encoding\":\"gzip\",\"user-agent\":\"Go-http-client/2.0\"},\"url\":\"https://postman-echo.com/get?id=32\"}"`),
			expect:         "32",
			pathExpression: "$.args.id",
		},
		{
			name:           "single depth unescaped",
			payload:        []byte(`{"args":{"id":"32"}}`),
			expect:         "32",
			pathExpression: "$.args.id",
		},
		{
			name: "lookup string in array ",
			payload: []byte(`{"store": {"book": [
				{"category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95},
				{"category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honour", "price": 12.99},
				{"category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "isbn": "0-553-21311-3", "price": 8.99},
				{"category": "fiction", "author": "J. R. R. Tolkien", "title": "The Lord of the Rings", "isbn": "0-395-19395-8", "price": 22.99}],
				"bicycle": {"color": "red", "price": 19.95}, "tools": null}}`),
			expect:         "The Lord of the Rings",
			pathExpression: "$.store.book[?(@.author=='J. R. R. Tolkien')].title",
		},
		{
			name:           "lookup int in array",
			payload:        []byte(`{"items":[{"id":3,"name":"fubar","a":"b","c":"d"},{"id":32,"name":"fubar2","a":"f","c":"h"},{"id":42,"name":"fubar42","a":"i","c":"j"}]}`),
			expect:         "3",
			pathExpression: "$.items[?(@.name=='fubar')].id",
		},
		{
			name: "lookup float in array ",
			payload: []byte(`{"store": {"book": [
				{"category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95},
				{"category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honour", "price": 12.99},
				{"category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "isbn": "0-553-21311-3", "price": 8.99},
				{"category": "fiction", "author": "J. R. R. Tolkien", "title": "The Lord of the Rings", "isbn": "0-395-19395-8", "price": 22.99}],
				"bicycle": {"color": "red", "price": 19.95}, "tools": null}}`),
			expect:         "22.99",
			pathExpression: "$.store.book[?(@.author=='J. R. R. Tolkien')].price",
		},
		{
			name:           "lookup object in array - expect error",
			payload:        []byte(`{"items":[{"id":3,"name":"fubar","object": {"f": "g"},"a":"b","c":"d"},{"id":32,"name":"fubar2","a":"f","c":"h"},{"id":42,"name":"fubar42","a":"i","c":"j"}]}`),
			expErr:         String(fmt.Sprintf("cannot use type: %v in further processing - can only be a numeric or string value", ajson.Object)),
			pathExpression: "$.items[?(@.name=='fubar')].object",
		},
		{
			name:           "lookup null in array - expect error",
			payload:        []byte(`{"items":[{"id":3,"name":"fubar","null": null,"a":"b","c":"d"},{"id":32,"name":"fubar2","a":"f","c":"h"},{"id":42,"name":"fubar42","a":"i","c":"j"}]}`),
			expErr:         String(fmt.Sprintf("cannot use type: %v in further processing - can only be a numeric or string value", ajson.Null)),
			pathExpression: "$.items[?(@.name=='fubar')].null",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SeederImpl{runtimeVars: runtimeVars{}, log: log.New(&bytes.Buffer{}, log.DebugLvl), client: &http.Client{}}
			got, err := r.findPathByExpression(tt.payload, tt.pathExpression)
			if err != nil {
				if tt.expErr != nil {
					if err.Error() != *tt.expErr {
						t.Errorf(testutils.TestPhrase, tt.expErr, err)
					}
				} else {
					t.Error(err)
				}
			}
			if tt.expErr != nil && err.Error() != *tt.expErr {
				t.Errorf(testutils.TestPhrase, tt.expErr, err)
			} else {
				if got != tt.expect {
					t.Errorf(testutils.TestPhrase, tt.expect, got)
				}
			}

		})
	}
}

func Test_templatePayload(t *testing.T) {
	tests := []struct {
		name      string
		rest      *SeederImpl
		payload   string
		variables map[string]any
		expect    string
	}{
		{
			name:      "global only",
			rest:      &SeederImpl{runtimeVars: runtimeVars{}, log: log.New(&bytes.Buffer{}, log.ErrorLvl), client: &http.Client{}},
			payload:   `{"foo":"${bar}","BAZ":"$FUZZ"}`,
			variables: map[string]any{},
			expect:    `{"foo":"","BAZ":"BOO"}`,
		},
		{
			name:      "global + injected",
			rest:      &SeederImpl{runtimeVars: runtimeVars{}, log: log.New(&bytes.Buffer{}, log.ErrorLvl), client: &http.Client{}},
			payload:   `{"foo":"${bar}","BAZ":"$FUZZ"}`,
			variables: map[string]any{"bar": "hoo"},
			expect:    `{"foo":"hoo","BAZ":"BOO"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("FUZZ", "BOO")

			got := tt.rest.templatePayload(tt.payload, tt.variables)
			if got != tt.expect {
				t.Errorf(testutils.TestPhrase, tt.expect, got)
			}
		})
	}
}

func Test_ActionWithHeader(t *testing.T) {
	tests := []struct {
		name   string
		action *Action
		header *map[string]string
		expect []string
	}{
		{
			name: "default header",
			action: &Action{
				Strategy:             "",
				Order:                new(int),
				Endpoint:             "",
				GetEndpointSuffix:    new(string),
				PostEndpointSuffix:   new(string),
				PutEndpointSuffix:    new(string),
				DeleteEndpointSuffix: new(string),
				FindByJsonPathExpr:   "",
				PayloadTemplate:      "",
				Variables:            map[string]any{},
				RuntimeVars:          &map[string]string{},
				AuthMapRef:           "",
				HttpHeaders:          &map[string]string{},
			},
			expect: []string{"Accept", "Content-Type"},
		},
		{
			name: "additional",
			action: &Action{
				Strategy:             "",
				Order:                new(int),
				Endpoint:             "",
				GetEndpointSuffix:    new(string),
				PostEndpointSuffix:   new(string),
				PutEndpointSuffix:    new(string),
				DeleteEndpointSuffix: new(string),
				FindByJsonPathExpr:   "",
				PayloadTemplate:      "",
				Variables:            map[string]any{},
				RuntimeVars:          &map[string]string{},
				AuthMapRef:           "",
				HttpHeaders:          &map[string]string{"X-Foo": "bar"},
			},
			expect: []string{"Accept", "Content-Type", "X-Foo"},
		},
		{
			name: "additional with custom overwrite",
			action: &Action{
				Strategy:             "",
				Order:                new(int),
				Endpoint:             "",
				GetEndpointSuffix:    new(string),
				PostEndpointSuffix:   new(string),
				PutEndpointSuffix:    new(string),
				DeleteEndpointSuffix: new(string),
				FindByJsonPathExpr:   "",
				PayloadTemplate:      "",
				Variables:            map[string]any{},
				RuntimeVars:          &map[string]string{},
				AuthMapRef:           "",
				HttpHeaders:          &map[string]string{"X-Foo": "bar", "Accept": "application/xml"},
			},
			expect: []string{"Accept", "Content-Type", "X-Foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			a := tt.action
			got := a.WithHeader()
			if got.header == nil {
				t.Error("failed to create local header on Action")
			}
			hc := 0
			for k := range *got.header {
				if !strings.Contains(fmt.Sprintf("%v", got.header), k) {
					t.Error("incorrect keys in header")
				}
				hc++
			}
			if hc != len(tt.expect) {
				t.Errorf("expected: %v, got: %v", len(tt.expect), hc)
			}
		})
	}
}

func Test_setRunTimeVars(t *testing.T) {

	tests := []struct {
		name                 string
		rest                 *SeederImpl
		createUpdateResponse []byte
		action               *Action
	}{
		{
			name:                 "vars found and replaced",
			rest:                 &SeederImpl{runtimeVars: runtimeVars{}, log: log.New(&bytes.Buffer{}, log.ErrorLvl)},
			createUpdateResponse: []byte(`{"id": "aaabbbccc"}`),
			action: &Action{
				name:                 "foo1",
				PayloadTemplate:      `{"foo": "${GLOBAL}","local": "${local}", "runtime":"${someId}" }`,
				PatchPayloadTemplate: "",
				Variables:            map[string]any{},
				RuntimeVars: &map[string]string{
					"someId": "$.id",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.rest.runtimeVars) > 0 {
				t.Errorf("runtimeVars should be empty at this point, instead found: %v", len(tt.rest.runtimeVars))
			}
			tt.rest.setRuntimeVar(tt.createUpdateResponse, tt.action)

			if len(tt.rest.runtimeVars) < 1 {
				t.Error("no vars found and replaced")
			}
		})
	}
}
