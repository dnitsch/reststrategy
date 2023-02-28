/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/errors"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kind/pkg/cluster"

	seederv1alpha1 "github.com/dnitsch/reststrategy/kubebuilder-controller/api/v1alpha1"
	"github.com/dnitsch/reststrategy/seeder"

	//+kubebuilder:scaffold:imports

	log "github.com/dnitsch/simplelog"
	ctrl "sigs.k8s.io/controller-runtime"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var k8sClient client.Client
var testEnv *envtest.Environment
var deleteCluster func()
var logr = log.NewLogr(os.Stdout, log.DebugLvl).V(1)
var logger = log.New(os.Stdout, log.DebugLvl)
var defaultClusterName string = "kubebuilder-test"

// ====
// BEGIN CUSTOM K8s setup
//
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

	// if len(os.Getenv("GITHUB_ACTIONS")) > 0 || len(os.Getenv("TRAVIS")) > 0 || len(os.Getenv("CIRCLECI")) > 0 || len(os.Getenv("GITLAB_CI")) > 0 {
	// 	fmt.Println("In CI will be using a service mounted KinD")
	// 	return func() {}
	// }

	usr, _ := user.Current()
	hd := usr.HomeDir
	kubeConfigPath := path.Join(hd, ".kube/config")
	// when Podman is available =>
	// KIND_EXPERIMENTAL_PROVIDER=podman kind create cluster --name kind-kind
	// or when using docker
	// kind create cluster --name kind-kind
	impProvider := detectContainerImp()
	if impProvider == nil {
		fmt.Println("unable to find suitable containerisation provider")
		// t.Errorf("unable to find suitable containerisation provider")
		// t.SkipNow()
		return func() {}
	}
	// logger := cmd.NewLogger()
	clusterProviderOptions := []cluster.ProviderOption{
		cluster.ProviderWithLogger(cmd.NewLogger()), //(log.New(os.Stdout, log.DebugLvl)), //logger.WithName("KinD set up")),
	}

	clusterProviderOptions = append(clusterProviderOptions, impProvider)

	provider := cluster.NewProvider(clusterProviderOptions...)
	// create the cluster
	if err := provider.Create(
		defaultClusterName,
		cluster.CreateWithNodeImage(""),
		cluster.CreateWithRetain(false),
		cluster.CreateWithWaitForReady(time.Second*60),
		cluster.CreateWithKubeconfigPath(""),
		cluster.CreateWithDisplayUsage(false),
		cluster.CreateWithDisplaySalutation(false),
		// cluster.
	); err != nil {
		fmt.Println("failed to create cluster")
		fmt.Println(err)
		t.Fatal(errors.Wrap(err, "failed to create cluster"))
	}
	return func() {
		// delete cluster
		if err := provider.Delete(defaultClusterName, kubeConfigPath); err != nil {
			t.Errorf("failed to tear down kind cluster: %s", err)
		}
	}
}

// k8s-client set up
func kubeClientSetup(t *testing.T) (*kubernetes.Clientset, *rest.Config, error) {
	usr, _ := user.Current()
	hd := usr.HomeDir
	kubeConfigPath := path.Join(hd, ".kube/config")

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialise client from config: %s", err.Error())
	}
	if b, err := os.ReadFile(kubeConfigPath); err != nil {
		logger.Errorf("kubeConfigPath err: %v", err)
	} else {
		logger.Infof("kubeConfigPath file (%s) contents: %v", kubeConfigPath, string(b))
	}
	logger.Infof("kubeConfigPath file (%s) yielded this config: %v", kubeConfigPath, cfg)

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("error building kubernetes clientset: %s", err.Error())
	}

	logger.Infof("config kubeClient: %v", kubeClient)
	return kubeClient, cfg, nil
}

//
// END CUSTOM K8s setup
// ====

func TestAPIs(t *testing.T) {

	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	t := &testing.T{}

	logf.SetLogger(logr.WithName("RestStrategyController-Test"))
	deleteCluster = startCluster(t)

	_, cfg, e := kubeClientSetup(t)
	if e != nil {
		t.Errorf("failed to get client: %v", e)
	}

	logger.Infof("config returned from kubeClient setup: %v", cfg)

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
		UseExistingCluster:    seeder.Bool(true),
		Config:                cfg,
	}

	_, err := testEnv.Start()

	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = seederv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})

	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})

	Expect(err).ToNot(HaveOccurred())

	err = (&RestStrategyReconciler{
		Client:       k8sManager.GetClient(),
		Scheme:       k8sManager.GetScheme(),
		ResyncPeriod: 1,
		Logger:       log.New(os.Stderr, log.DebugLvl),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(context.TODO())
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	if deleteCluster != nil {
		deleteCluster()
	}
	Expect(err).NotTo(HaveOccurred())
})
