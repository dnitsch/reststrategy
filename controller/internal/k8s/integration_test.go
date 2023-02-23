package k8s_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	clientset "github.com/dnitsch/reststrategy/apis/reststrategy/generated/clientset/versioned"
	"github.com/dnitsch/reststrategy/apis/reststrategy/v1alpha1"
	"github.com/dnitsch/reststrategy/controller/internal/k8s"
	log "github.com/dnitsch/simplelog"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/errors"
	"sigs.k8s.io/yaml"
)

var defaultClusterName string = "int"
var defaultCrdNs string = "runtime-config-sync-in-k8s"

// detect either podman or docker
func detectContainerImp() cluster.ProviderOption {
	if imp, ok := os.LookupEnv("DOCKER_HOST"); ok {
		docker, podman := strings.HasSuffix(imp, "docker.sock"), strings.HasSuffix(imp, "podman.sock")
		if docker {
			return cluster.ProviderWithDocker()
		}
		if podman {
			return cluster.ProviderWithPodman()
		}
	}
	return nil
}

// start kind cluster
func startCluster(t *testing.T) func() {
	// when Podman is available =>
	// KIND_EXPERIMENTAL_PROVIDER=podman kind create cluster --name kind-kind
	// or when using docker
	// kind create cluster --name kind-kind
	impProvider := detectContainerImp()
	if impProvider == nil {
		t.Errorf("unable to find suitable containerisation provider")
		t.SkipNow()
		return func() {}
	}

	// logger := log.New(&bytes.Buffer{}, log.DebugLvl)
	clusterProviderOptions := []cluster.ProviderOption{
		cluster.ProviderWithLogger(cmd.NewLogger()),
	}

	clusterProviderOptions = append(clusterProviderOptions, impProvider)

	provider := cluster.NewProvider(clusterProviderOptions...)
	// create the cluster
	if err := provider.Create(
		defaultClusterName,
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
		t.Logf("called defer function")
		if err := provider.Delete(defaultClusterName, ""); err != nil {
			t.Errorf("failed to tear down kind cluster: %s", err)
		}
	}
}

// k8s-client set up
func kubeClientSetup(t *testing.T) (*kubernetes.Clientset, *rest.Config, error) {
	usr, _ := user.Current()
	hd := usr.HomeDir
	cfg, err := clientcmd.BuildConfigFromFlags("", path.Join(hd, ".kube/config"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialise client from config: %s", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("error building kubernetes clientset: %s", err.Error())
	}
	return kubeClient, cfg, nil
}

// create NS
func createNameSpace(t *testing.T, log log.Loggeriface, kubeClient *kubernetes.Clientset) (func(), error) {
	ns := v1.Namespace(defaultCrdNs)
	if err := yaml.Unmarshal([]byte(fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    role: test-int-in-k8s`, defaultCrdNs)), ns); err != nil {
		log.Errorf("failed to create ns: %s", err.Error())
		t.Fatal()
	}

	if _, err := kubeClient.CoreV1().Namespaces().Apply(context.TODO(), ns, metav1.ApplyOptions{FieldManager: "application/apply-patch"}); err != nil {
		t.Error(err)
		return nil, err
	}
	return func() {
		if err := kubeClient.CoreV1().Namespaces().Delete(context.TODO(), defaultCrdNs, metav1.DeleteOptions{}); err != nil {
			t.Errorf("failed to delete ns: %s", err.Error())
		}
	}, nil

}

// CreateCRDSchema only needs to be done once
func createCRDRestStrategy(log log.Loggeriface, cfg *rest.Config) (func(), error) {
	// k8s-client apply CRD
	extClient, _ := apiextension.NewForConfig(cfg)
	_, currentFile, _, _ := runtime.Caller(0)
	crd := &apiextensionv1.CustomResourceDefinition{}
	// crd dir
	crdSchemaFile := path.Join(path.Dir(path.Base(currentFile)), "../../../", "crd", "reststrategy.yml")
	crdB, _ := os.ReadFile(crdSchemaFile)

	if err := yaml.Unmarshal(crdB, crd); err != nil {
		return nil, err
	}

	crd, err := extClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), crd, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return func() {
		_ = extClient.ApiextensionsV1().CustomResourceDefinitions().Delete(context.TODO(), crd.Name, metav1.DeleteOptions{})
	}, nil
}

func crdClientSetup(t *testing.T, log log.Loggeriface, cfg *rest.Config) (*clientset.Clientset, error) {
	return clientset.NewForConfig(cfg)
}

// applyRestStrategy called from each test case
func applyRestStrategy(t *testing.T, log log.Loggeriface, reststrategyClient *clientset.Clientset) (*v1alpha1.RestStrategy, func(), error) {
	// test file location
	_, currentFile, _, _ := runtime.Caller(0)
	testCrd := &v1alpha1.RestStrategy{}
	test1File := path.Join(path.Dir(path.Base(currentFile)), "../../../", "test", "integration-test1.yml")
	b, err := os.ReadFile(test1File)
	if err != nil {
		t.Fatal()
	}

	if err := yaml.Unmarshal(b, testCrd); err != nil {
		t.Fatal()
	}

	rst, err := reststrategyClient.ReststrategyV1alpha1().RestStrategies(defaultCrdNs).Create(context.TODO(), testCrd, metav1.CreateOptions{})

	return rst, func() {
		if err := reststrategyClient.ReststrategyV1alpha1().RestStrategies(defaultCrdNs).Delete(context.TODO(), "test-int-local", metav1.DeleteOptions{}); err != nil {
			t.Errorf("failed to delete ns: %s", err.Error())
		}
	}, err
}

// Only debug this test as run in VSCode will always time out
func TestIntegration(t *testing.T) {
	t.SkipNow()
	t.Setenv("test.timeout", "30m0s")
	flag.Set("test.timeout", "30m0s")
	logger := log.New(&bytes.Buffer{}, log.DebugLvl)
	// beforeAll as it shouldn't be countred
	// create cluster
	// defer delete after all
	// ENABLE once tested
	// ==>
	deleteCluster := startCluster(t)
	defer deleteCluster()
	// <===
	// ENABLE once tested
	kubeclient, cfg, err := kubeClientSetup(t)
	if err != nil {
		t.FailNow()
	}
	stopCh := make(chan struct{})
	customClient, err := crdClientSetup(t, logger, cfg)
	if err != nil {
		t.Fatal(err)
	}

	go func(stopCh chan struct{}) {
		if err := k8s.RunWithConfig(k8s.Config{Namespace: defaultCrdNs}, logger, stopCh, kubeclient, customClient); err != nil {
			t.Fatal("failed to start Controller")
		}
	}(stopCh)

	ttests := map[string]struct {
		logger       log.Loggeriface
		cfg          *rest.Config
		kubeclient   *kubernetes.Clientset
		config       k8s.Config
		expectStatus string
	}{
		"success test1": {
			logger:       logger,
			cfg:          cfg,
			kubeclient:   kubeclient,
			config:       k8s.Config{Namespace: defaultCrdNs},
			expectStatus: "RestStrategy successfully executed",
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			t.SkipNow()
			deleteNs, err := createNameSpace(t, tt.logger, tt.kubeclient)
			if err != nil {
				t.Fatal(err)
			}
			defer deleteNs()

			deleteCrd, err := createCRDRestStrategy(tt.logger, cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer deleteCrd()

			crd, crdDelete, err := applyRestStrategy(t, tt.logger, customClient)
			if err != nil {
				t.Error(err)
			}
			defer crdDelete()
			if crd == nil {
				t.Error("expected not nil")
			}
			if crd.Status.Message != tt.expectStatus {
				t.Errorf("status mesage is not expected got: %s\n\nwant: %s", crd.Status.Message, tt.expectStatus)
			}
		})
	}
	close(stopCh)
}
