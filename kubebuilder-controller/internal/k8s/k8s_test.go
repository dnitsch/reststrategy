package k8s_test

import (
	"os"
	"testing"

	"github.com/dnitsch/reststrategy/kubebuilder-controller/internal/k8s"
)

func TestGetNamespce(t *testing.T) {
	ttests := map[string]struct {
		setup  func(t *testing.T) func()
		input  string
		expect string
	}{
		"use env var": {
			func(t *testing.T) func() {
				os.Setenv("POD_NAMESPACE", "rand-ns")
				return func() {
					os.Clearenv()
				}
			},
			"",
			"rand-ns",
		},
		"use config var over env var": {
			func(t *testing.T) func() {
				os.Setenv("POD_NAMESPACE", "rand-ns")
				return func() {
					os.Clearenv()
				}
			},
			"config-passed-ns",
			"config-passed-ns",
		},
		"neither supplied": {
			func(t *testing.T) func() {
				return func() {
				}
			},
			"",
			"",
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			clear := tt.setup(t)
			defer clear()
			if got := k8s.GetNamespace(tt.input); got != tt.expect {
				t.Errorf("\ngot: %s\n\nwanted: %s\n\n", got, tt.expect)
			}
		})
	}
}
