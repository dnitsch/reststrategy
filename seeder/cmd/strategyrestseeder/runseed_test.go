package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnitsch/reststrategy/seeder"

	"gopkg.in/yaml.v2"
)

func helperTestSeed(conf *seeder.StrategyConfig) string {
	b, _ := yaml.Marshal(conf)
	dir, _ := os.MkdirTemp("", "seed-cli-test")
	file := filepath.Join(dir, "seeder.yml")
	_ = os.WriteFile(file, b, 0777)
	return file
}

func TestMainRunSeed(t *testing.T) {
	ttests := map[string]struct {
		// path to file and delete file return
		testInput func(t *testing.T, url string) ([]string, func())
		expect    string
	}{
		"no auth and no seed": {
			func(t *testing.T, url string) ([]string, func()) {
				conf := &seeder.StrategyConfig{
					AuthConfig: seeder.AuthMap{},
					Seeders:    seeder.Seeders{},
				}
				file := helperTestSeed(conf)
				return []string{"run", "-p", file}, func() {
					os.Remove(file)
				}
			},
			"",
		},
		"verbose no auth no seed": {
			func(t *testing.T, url string) ([]string, func()) {
				conf := &seeder.StrategyConfig{
					AuthConfig: seeder.AuthMap{},
					Seeders:    seeder.Seeders{},
				}
				file := helperTestSeed(conf)
				return []string{"run", "-p", file, "-v"}, func() {
					os.Remove(file)
				}
			},
			"",
		},
		"with configmanager no auth no seed": {
			func(t *testing.T, url string) ([]string, func()) {
				conf := &seeder.StrategyConfig{
					AuthConfig: seeder.AuthMap{},
					Seeders:    seeder.Seeders{},
				}
				file := helperTestSeed(conf)
				return []string{"run", "-p", file, "-v", "-c", "-t", "://", "-k", "|"}, func() {
					os.Remove(file)
				}
			},
			"",
		},
		"no args supplied": {
			func(t *testing.T, url string) ([]string, func()) {
				conf := &seeder.StrategyConfig{
					AuthConfig: seeder.AuthMap{},
					Seeders:    seeder.Seeders{},
				}
				file := helperTestSeed(conf)
				return []string{"run"}, func() {
					os.Remove(file)
				}
			},
			"must include input",
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			cmdArgs, cleanUp := tt.testInput(t, "")
			defer cleanUp()
			b := new(bytes.Buffer)

			cmd := StrategyRestSeederCmd

			cmd.SetArgs(cmdArgs)
			cmd.SetErr(b)
			cmd.Execute()
			out, err := io.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			//
			if tt.expect == "" && len(out) > 0 {
				t.Errorf(`%s 
got: %v
wanted: ""`, "expected empty buffer", string(out))
			}

			if tt.expect != "" && !strings.Contains(string(out), tt.expect) {
				t.Errorf(`%s 
got: %v
want: %v`, "output comparison failed", string(out), tt.expect)
			}

			cmd = nil
			path = ""
			cmKeySeparator = ""
			cmTokenSeparator = ""
			verbose = false
		})
	}
}
