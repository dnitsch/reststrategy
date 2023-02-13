package seeder_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/dnitsch/reststrategy/seeder"
	log "github.com/dnitsch/simplelog"
)

func Test_findByPathExpression(t *testing.T) {
	tests := map[string]struct {
		payload        []byte
		pathExpression string
		expect         string
	}{
		"single depth escaped": {
			payload:        []byte(`"{\"args\":{\"id\":\"32\"},\"headers\":{\"x-forwarded-proto\":\"https\",\"x-forwarded-port\":\"443\",\"host\":\"postman-echo.com\",\"x-amzn-trace-id\":\"Root=1-63106cc4-6e8b66b055d278e5613db058\",\"content-length\":\"2\",\"accept\":\"application/json\",\"content-type\":\"application/json\",\"accept-encoding\":\"gzip\",\"user-agent\":\"Go-http-client/2.0\"},\"url\":\"https://postman-echo.com/get?id=32\"}"`),
			expect:         "32",
			pathExpression: "$.args.id",
		},
		"single depth unescaped": {
			payload:        []byte(`{"args":{"id":"32"}}`),
			expect:         "32",
			pathExpression: "$.args.id",
		},
		"lookup string in array ": {
			payload: []byte(`{"store": {"book": [
				{"category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95},
				{"category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honour", "price": 12.99},
				{"category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "isbn": "0-553-21311-3", "price": 8.99},
				{"category": "fiction", "author": "J. R. R. Tolkien", "title": "The Lord of the Rings", "isbn": "0-395-19395-8", "price": 22.99}],
				"bicycle": {"color": "red", "price": 19.95}, "tools": null}}`),
			expect:         "The Lord of the Rings",
			pathExpression: "$.store.book[?(@.author=='J. R. R. Tolkien')].title",
		},
		"lookup int in array": {
			payload:        []byte(`{"items":[{"id":3,"name":"fubar","a":"b","c":"d"},{"id":32,"name":"fubar2","a":"f","c":"h"},{"id":42,"name":"fubar42","a":"i","c":"j"}]}`),
			expect:         "3",
			pathExpression: "$.items[?(@.name=='fubar')].id",
		},
		"lookup float in array ": {
			payload: []byte(`{"store": {"book": [
				{"category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95},
				{"category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honour", "price": 12.99},
				{"category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "isbn": "0-553-21311-3", "price": 8.99},
				{"category": "fiction", "author": "J. R. R. Tolkien", "title": "The Lord of the Rings", "isbn": "0-395-19395-8", "price": 22.99}],
				"bicycle": {"color": "red", "price": 19.95}, "tools": null}}`),
			expect:         "22.99",
			pathExpression: "$.store.book[?(@.author=='J. R. R. Tolkien')].price",
		},
		"lookup object in array - expect error": {
			payload:        []byte(`{"items":[{"id":3,"name":"fubar","object": {"f": "g"},"a":"b","c":"d"},{"id":32,"name":"fubar2","a":"f","c":"h"},{"id":42,"name":"fubar42","a":"i","c":"j"}]}`),
			expect:         "cannot use type: 5 in further processing - can only be a numeric or string value",
			pathExpression: "$.items[?(@.name=='fubar')].object",
		},
		"lookup null in array - expect error": {
			payload:        []byte(`{"items":[{"id":3,"name":"fubar","null": null,"a":"b","c":"d"},{"id":32,"name":"fubar2","a":"f","c":"h"},{"id":42,"name":"fubar42","a":"i","c":"j"}]}`),
			expect:         "cannot use type: 0 in further processing - can only be a numeric or string value",
			pathExpression: "$.items[?(@.name=='fubar')].null",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := seeder.NewSeederImpl(log.New(&bytes.Buffer{}, log.ErrorLvl))
			got, err := r.FindPathByExpression(tt.payload, tt.pathExpression)
			if err != nil {
				if err.Error() != tt.expect {
					t.Errorf(TestPhrase, err, tt.expect)
				}
				return
			}

			if got != tt.expect {
				t.Errorf(TestPhrase, got, tt.expect)
			}

		})
	}
}

func Test_templatePayload(t *testing.T) {
	tests := []struct {
		name      string
		rest      *seeder.SeederImpl
		payload   string
		variables map[string]any
		expect    string
	}{
		{
			name:      "global only",
			rest:      &seeder.SeederImpl{},
			payload:   `{"foo":"${bar}","BAZ":"$FUZZ"}`,
			variables: map[string]any{},
			expect:    `{"foo":"","BAZ":"BOO"}`,
		},
		{
			name:      "global + injected",
			rest:      &seeder.SeederImpl{},
			payload:   `{"foo":"${bar}","BAZ":"$FUZZ"}`,
			variables: map[string]any{"bar": "hoo"},
			expect:    `{"foo":"hoo","BAZ":"BOO"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("FUZZ", "BOO")

			got := tt.rest.TemplatePayload(tt.payload, tt.variables)
			if got != tt.expect {
				t.Errorf(TestPhrase, got, tt.expect)
			}
		})
	}
}

func Test_ActionWithHeader(t *testing.T) {
	tests := []struct {
		name   string
		action *seeder.Action
		header *map[string]string
		expect []string
	}{
		{
			name: "default header",
			action: &seeder.Action{
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
				RuntimeVars:          map[string]string{},
				AuthMapRef:           "",
				HttpHeaders:          &map[string]string{},
			},
			expect: []string{"Accept", "Content-Type"},
		},
		{
			name: "additional",
			action: &seeder.Action{
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
				RuntimeVars:          map[string]string{},
				AuthMapRef:           "",
				HttpHeaders:          &map[string]string{"X-Foo": "bar"},
			},
			expect: []string{"Accept", "Content-Type", "X-Foo"},
		},
		{
			name: "additional with custom overwrite",
			action: &seeder.Action{
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
				RuntimeVars:          map[string]string{},
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
			if got.HttpHeaders == nil {
				t.Error("failed to create local header on Action")
			}
			// hc := 0
			// for k := range *got.header {
			// 	if !strings.Contains(fmt.Sprintf("%v", got.header), k) {
			// 		t.Error("incorrect keys in header")
			// 	}
			// 	hc++
			// }
			// if hc != len(tt.expect) {
			// 	t.Errorf("expected: %v, got: %v", len(tt.expect), hc)
			// }
		})
	}
}

// Tested implicitely
// func Test_setRunTimeVars(t *testing.T) {

// 	tests := map[string]struct {
// 		rest                 *seeder.SeederImpl
// 		createUpdateResponse []byte
// 		action               *seeder.Action
// 	}{
// 		"vars found and replaced": {
// 			rest:                 &seeder.SeederImpl{},
// 			createUpdateResponse: []byte(`{"id": "aaabbbccc"}`),
// 			action: &seeder.Action{
// 				// name:                 "foo1",
// 				PayloadTemplate:      `{"foo": "${GLOBAL}","local": "${local}", "runtime":"${someId}" }`,
// 				PatchPayloadTemplate: "",
// 				Variables:            map[string]any{},
// 				RuntimeVars: map[string]string{
// 					"someId": "$.id",
// 				},
// 			},
// 		},
// 	}
// 	for name, tt := range tests {
// 		t.Run(name, func(t *testing.T) {

// 			if len(tt.seeder.RuntimeVars()) > 0 {
// 				t.Errorf("runtimeVars should be empty at this point, instead found: %v", len(tt.seeder.RuntimeVars()))
// 			}
// 			tt.seeder.SetRuntimeVar(tt.createUpdateResponse, tt.action)

// 			if len(tt.seeder.runtimeVars) < 1 {
// 				t.Error("no vars found and replaced")
// 			}
// 		})
// 	}
// }
