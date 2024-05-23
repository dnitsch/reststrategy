package controller

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

func helperFileSeeder(conf *seeder.StrategyConfig) (string, string) {
	b, _ := yaml.Marshal(conf)
	dir, _ := os.MkdirTemp("", "seed-cli-test")
	file := filepath.Join(dir, "seeder.yml")
	k8sConf := filepath.Join(dir, ".kube-config")
	_ = os.WriteFile(file, b, 0777)
	_ = os.WriteFile(k8sConf, []byte(`
apiVersion: v1
kind: Config
preferences: {}`), 0777)
	return k8sConf, file
}

func TestCmdController(t *testing.T) {
	t.Skip()
	ttests := map[string]struct {
		// path to file and delete file return
		testInput func(t *testing.T, url string) ([]string, func())
		expect    string
	}{
		"no config manager": {
			func(t *testing.T, url string) ([]string, func()) {
				kubeConf, sourceFile := helperFileSeeder(&seeder.StrategyConfig{})
				return []string{"--kubeconfig", kubeConf}, func() {
					masterURL = ""
					kubeconfig = ""
					rsyncperiod = 0
					namespace = ""
					logLevel = ""
					metricsAddr = ""
					enableLeaderElection = false
					enableConfigManager = false
					configManagerTokenSeparator = ""
					configManagerKeySeparator = ""
					probeAddr = ""
					os.Remove(sourceFile)
					os.Remove(kubeConf)
				}
			},
			"",
		},
		"with config manager": {
			func(t *testing.T, url string) ([]string, func()) {

				kubeConf, sourceFile := helperFileSeeder(&seeder.StrategyConfig{})
				return []string{"-c", "--kubeconfig", kubeConf}, func() {
					masterURL = ""
					kubeconfig = ""
					rsyncperiod = 0
					namespace = ""
					logLevel = ""
					metricsAddr = ""
					enableLeaderElection = false
					enableConfigManager = false
					configManagerTokenSeparator = ""
					configManagerKeySeparator = ""
					probeAddr = ""
					os.Remove(sourceFile)
					os.Remove(kubeConf)
				}
			},
			"",
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			cmdArgs, cleanUp := tt.testInput(t, "")
			defer cleanUp()
			b := new(bytes.Buffer)

			cmd := ControllerCmd

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

		})
	}
}
