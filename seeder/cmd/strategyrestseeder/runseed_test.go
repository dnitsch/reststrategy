package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/dnitsch/reststrategy/seeder"
	"gopkg.in/yaml.v2"
	// cmd "github.com/dnitsch/reststrategy/seeder/cmd/strategyrestseeder"
)

func helperTestSeed(conf *seeder.StrategyConfig) string {
	originalArg := os.Args[0:1]
	os.Args = originalArg

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
			b := &bytes.Buffer{}
			cmd := runCmd
			// cmd.SetArgs did not work with this set up
			os.Args = os.Args[0:1]
			os.Args = append(os.Args, cmdArgs...)
			cmd.SetOut(b)
			cmd.Execute()
			out, err := io.ReadAll(b)
			if err != nil {
				if err.Error() != tt.expect {
					t.Fatal(err)
				}
				return
			}
			if string(out) != tt.expect {
				t.Errorf(`%s 
got: %v
want: %v`, "output comparison failed", string(out), tt.expect)
			}

		})
	}
}
