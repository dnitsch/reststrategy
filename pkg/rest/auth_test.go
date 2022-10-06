package rest

import (
	"testing"

	"github.com/dnitsch/strategyrestseeder/internal/testutils"
)

func Test_InitAuthMap(t *testing.T) {
	tests := []struct {
		name    string
		authMap *AuthMap
	}{
		{
			name: "default custom vals",
			authMap: &AuthMap{
				"customTest1": {
					AuthStrategy: CustomToToken,
					CustomToken: &CustomToken{
						AuthUrl:       "https://foo.bar",
						CustomAuthMap: map[string]any{"name": "skywalker", "secret_pass": "empire"},
						HeaderKey:     "Foo",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			a := NewAuth(tt.authMap)
			for _, d := range *a {
				if d.authStrategy != CustomToToken {
					t.Errorf("incorrect authStrategy: %v, wanted: %s", d.authStrategy, "CustomToToken")
				}
				if d.customToToken.headerKey != "Foo" {
					t.Errorf(testutils.TestPhrase, "Foo", d.customToToken.headerKey)
				}
			}
		})
	}
}
