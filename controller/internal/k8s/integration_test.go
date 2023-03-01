package k8s_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	clientset "github.com/dnitsch/reststrategy/apis/reststrategy/generated/clientset/versioned"
	"github.com/dnitsch/reststrategy/apis/reststrategy/v1alpha1"
	"github.com/dnitsch/reststrategy/controller/internal/k8s"
	cluster "github.com/dnitsch/reststrategy/kube-testing-tools/pkg/cluster"
	log "github.com/dnitsch/simplelog"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/yaml"
)

var defaultClusterName string = "int"
var defaultCrdNs string = "runtime-config-sync-in-k8s"

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
	tcluster := &cluster.ClusterTester{Name: defaultClusterName, EnsureUp: true}
	kubeStartUpConfig := tcluster.DetermineKubeConfig()

	deleteCluster := tcluster.StartCluster(t, kubeStartUpConfig)
	defer deleteCluster()

	kubeclient, cfg, err := tcluster.KubeClientSetup(t, kubeStartUpConfig)
	if err != nil {
		t.FailNow()
	}

	// <===
	// ENABLE once tested
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
