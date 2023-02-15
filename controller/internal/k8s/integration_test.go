package k8s_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	log "github.com/dnitsch/simplelog"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/errors"
)

// GetDefault selected the default runtime from the environment override
func GetDefault(logger log.Logger) cluster.ProviderOption {
	switch p := os.Getenv("KIND_EXPERIMENTAL_PROVIDER"); p {
	case "":
		return nil
	case "podman":
		logger.Info("using podman due to KIND_EXPERIMENTAL_PROVIDER")
		return cluster.ProviderWithPodman()
	case "docker":
		logger.Info("using docker due to KIND_EXPERIMENTAL_PROVIDER")
		return cluster.ProviderWithDocker()
	default:
		logger.Infof("ignoring unknown value %q for KIND_EXPERIMENTAL_PROVIDER", p)
		return nil
	}
}

// detect either podman or docker
func detectContainerImp() string {
	if imp, ok := os.LookupEnv("DOCKER_HOST"); ok {
		return imp
	}
	return ""
}

// start kind cluster
func startCluster(t *testing.T) func() {
	// when Podman is available =>
	// KIND_EXPERIMENTAL_PROVIDER=podman kind create cluster --name kind-kind
	// or when using docker
	// kind create cluster --name kind-kind
	impProvider := detectContainerImp()
	if impProvider == "" {
		t.FailNow()
		return func() {}
	}

	logger := log.New(&bytes.Buffer{}, log.DebugLvl)
	clusterProviderOptions := []cluster.ProviderOption{
		cluster.ProviderWithLogger(cmd.NewLogger()),
		GetDefault(logger),
	}

	clusterProviderOptions = append(clusterProviderOptions, cluster.ProviderWithPodman())

	provider := cluster.NewProvider(clusterProviderOptions...)
	// create the cluster
	if err := provider.Create(
		"kind-integration",
		// withConfig,
		cluster.CreateWithNodeImage(""),
		cluster.CreateWithRetain(false),
		cluster.CreateWithWaitForReady(time.Second*30),
		cluster.CreateWithKubeconfigPath(""),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create cluster"))
	}
	return func() {
		// delete cluster
		if err := provider.Delete("kind-integration", ""); err != nil {
			t.Errorf("failed to tear down kind cluster: %s", err)
		}
	}
}

// k8s-client apply CRD Schema
// k8s-client apply test crd
//

func TestIntegration(t *testing.T) {
	// beforeAll
	// create cluster
	// defer delete after all
	deleteCluster := startCluster(t)
	defer deleteCluster()
	ttests := map[string]struct {
		objType any
	}{
		"test1": {
			objType: nil,
		},
	}
	for name, tt := range ttests {

		t.Run(name, func(t *testing.T) {
			_ = tt
		})
	}
}
