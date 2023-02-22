package cmd

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/dnitsch/reststrategy/seeder"
	"gopkg.in/yaml.v2"
	// cmd "github.com/dnitsch/reststrategy/seeder/cmd/strategyrestseeder"
)

func TestMain(t *testing.T) {
	ttests := map[string]struct {
		// path to file and delete file return
		testInput func(t *testing.T, url string) (string, func())
		expect    string
	}{
		"no auth and no seed": {
			func(t *testing.T, url string) (string, func()) {
				conf := &seeder.StrategyConfig{
					AuthConfig: seeder.AuthMap{},
					Seeders:    seeder.Seeders{},
				}
				b, _ := yaml.Marshal(conf)
				dir, _ := os.MkdirTemp("", "seed-cli-test")
				file := filepath.Join(dir, "empty.yml")
				_ = os.WriteFile(file, b, fs.FileMode(os.O_RDONLY))
				return file, func() {
					os.Remove(file)
				}
			},
			"",
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			file, cleanUp := tt.testInput(t, "")
			defer cleanUp()
			b := &bytes.Buffer{}
			cmd := runCmd
			cmd.SetOut(b)
			cmd.SetArgs([]string{"run", "-p", file})
			cmd.Execute()
			out, err := io.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			if string(out) != tt.expect {
				t.Errorf(`%s 
got: %v, 
want: %v`, "string comparison failed", string(out), tt.expect)
			}

		})
	}
}
