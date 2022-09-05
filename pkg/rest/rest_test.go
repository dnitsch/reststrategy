package rest

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	log "github.com/dnitsch/simplelog"
	"github.com/dnitsch/strategyrestseeder/internal/testutils"
	"github.com/spyzhov/ajson"
)

func Test_getSeeder(t *testing.T) {
	tests := []struct {
		name   string
		action *Action
		rimpl  *SeederImpl
		expect string
	}{
		{
			name:  "getRestFunc",
			rimpl: &SeederImpl{client: &http.Client{}, log: log.New(os.Stdout, log.DebugLvl)},
			action: &Action{
				PayloadTemplate:    "{}",
				Strategy:           "GET/POST",
				Endpoint:           "https://postman-echo.com/get?id=32",
				FindByJsonPathExpr: "$.args.id",
			},
			expect: "32",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.rimpl.GetPost(context.TODO(), tt.action)
			if err != nil {
				t.Errorf("failed %s", err)
			}
			// if v != tt.expect {
			// 	t.Errorf(testutils.TestPhrase, tt.expect, v)
			// }
		})
	}
}

func Test_findByPathExpression(t *testing.T) {
	tests := []struct {
		name           string
		payload        []byte
		pathExpression string
		expect         string
		expErr         bool
	}{
		{
			name:           "single depth escaped",
			payload:        []byte(`"{\"args\":{\"id\":\"32\"},\"headers\":{\"x-forwarded-proto\":\"https\",\"x-forwarded-port\":\"443\",\"host\":\"postman-echo.com\",\"x-amzn-trace-id\":\"Root=1-63106cc4-6e8b66b055d278e5613db058\",\"content-length\":\"2\",\"accept\":\"application/json\",\"content-type\":\"application/json\",\"accept-encoding\":\"gzip\",\"user-agent\":\"Go-http-client/2.0\"},\"url\":\"https://postman-echo.com/get?id=32\"}"`),
			expect:         "32",
			expErr:         false,
			pathExpression: "$.args.id",
		},
		{
			name:           "single depth unescaped",
			payload:        []byte(`{"args":{"id":"32"}}`),
			expErr:         false,
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
			expErr:         false,
			expect:         "The Lord of the Rings",
			pathExpression: "$.store.book[?(@.author=='J. R. R. Tolkien')].title",
		},
		{
			name:           "lookup int in array",
			payload:        []byte(`{"items":[{"id":3,"name":"fubar","a":"b","c":"d"},{"id":32,"name":"fubar2","a":"f","c":"h"},{"id":42,"name":"fubar42","a":"i","c":"j"}]}`),
			expErr:         false,
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
			expErr:         false,
			expect:         "22.99",
			pathExpression: "$.store.book[?(@.author=='J. R. R. Tolkien')].price",
		},
		{
			name:           "lookup object in array - expect error",
			payload:        []byte(`{"items":[{"id":3,"name":"fubar","object": {"f": "g"},"a":"b","c":"d"},{"id":32,"name":"fubar2","a":"f","c":"h"},{"id":42,"name":"fubar42","a":"i","c":"j"}]}`),
			expErr:         true,
			expect:         fmt.Sprintf("cannot use type: %v in further processing - can only be a numeric or string value", ajson.Object),
			pathExpression: "$.items[?(@.name=='fubar')].object",
		},
		{
			name:           "lookup null in array - expect error",
			payload:        []byte(`{"items":[{"id":3,"name":"fubar","null": null,"a":"b","c":"d"},{"id":32,"name":"fubar2","a":"f","c":"h"},{"id":42,"name":"fubar42","a":"i","c":"j"}]}`),
			expErr:         true,
			expect:         fmt.Sprintf("cannot use type: %v in further processing - can only be a numeric or string value", ajson.Null),
			pathExpression: "$.items[?(@.name=='fubar')].null",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SeederImpl{log: log.New(os.Stderr, log.DebugLvl), client: &http.Client{}}
			got, err := r.findPathByExpression(tt.payload, tt.pathExpression)
			if err != nil {
				if tt.expErr {
					if err.Error() != tt.expect {
						t.Error(err)
					}
				} else {
					t.Error(err)
				}
			}
			if !tt.expErr && got != tt.expect {
				t.Errorf(testutils.TestPhrase, tt.expect, got)
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
			rest:      &SeederImpl{log: log.New(os.Stderr, log.DebugLvl), client: &http.Client{}},
			payload:   `{"foo":"${bar}","BAZ":"$FUZZ"}`,
			variables: map[string]any{},
			expect:    `{"foo":"","BAZ":"BOO"}`,
		},
		{
			name:      "global + injected",
			rest:      &SeederImpl{log: log.New(os.Stderr, log.DebugLvl), client: &http.Client{}},
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
