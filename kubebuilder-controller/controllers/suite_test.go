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

	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/errors"

	podv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
var kubeStartUpConfig kubeConfig = kubeConfig{}

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

type kubeConfig struct {
	isInCI        bool
	k8sConfigPath string
	kindConfig    *v1alpha4.Cluster
	masterUrl     string
}

func DetermineKubeConfig() kubeConfig {
	usr, _ := user.Current()
	hd := usr.HomeDir
	kubeConfigPath := path.Join(hd, ".kube/config")
	resp := kubeConfig{isInCI: false, k8sConfigPath: kubeConfigPath, kindConfig: nil, masterUrl: ""}
	if len(os.Getenv("GITHUB_ACTIONS")) > 0 ||
		len(os.Getenv("TRAVIS")) > 0 ||
		len(os.Getenv("CIRCLECI")) > 0 ||
		len(os.Getenv("GITLAB_CI")) > 0 ||
		len(os.Getenv("CI")) > 0 {
		logger.Info("In CI will be using a service mounted KinD")
		resp.kindConfig = &v1alpha4.Cluster{
			Networking: v1alpha4.Networking{
				APIServerAddress: "127.0.0.1",
				APIServerPort:    6443,
			},
		}
		// resp.masterUrl = "https://127.0.0.1:6443"
		return resp
	}
	return resp
}

// start kind cluster
func startCluster(t *testing.T) func() {
	impProvider := detectContainerImp()
	if impProvider == nil {
		logger.Info("unable to find suitable containerisation provider")
		return func() {}
	}
	// logger := cmd.NewLogger()
	clusterProviderOptions := []cluster.ProviderOption{
		cluster.ProviderWithLogger(cmd.NewLogger()),
	}

	clusterProviderOptions = append(clusterProviderOptions, impProvider)

	provider := cluster.NewProvider(clusterProviderOptions...)
	clusterCreateOptions := []cluster.CreateOption{
		cluster.CreateWithNodeImage(""),
		cluster.CreateWithRetain(false),
		cluster.CreateWithWaitForReady(time.Second * 60),
		cluster.CreateWithKubeconfigPath(kubeStartUpConfig.k8sConfigPath),
		cluster.CreateWithDisplayUsage(false),
		cluster.CreateWithDisplaySalutation(false),
	}

	if kubeStartUpConfig.isInCI && kubeStartUpConfig.kindConfig != nil {
		clusterCreateOptions = append(clusterCreateOptions, cluster.CreateWithV1Alpha4Config(kubeStartUpConfig.kindConfig))
	}
	// create the cluster
	if err := provider.Create(
		defaultClusterName,
		clusterCreateOptions...,
	); err != nil {
		logger.Errorf("failed to create cluster: %v", err)
		t.Fatal(errors.Wrap(err, "failed to create cluster"))
	}
	return func() {
		// delete cluster
		if err := provider.Delete(defaultClusterName, kubeStartUpConfig.k8sConfigPath); err != nil {
			t.Errorf("failed to tear down kind cluster: %s", err)
		}
	}
}

// k8s-client set up
func kubeClientSetup(t *testing.T) (*kubernetes.Clientset, *rest.Config, error) {

	// grab the internal IP and pass that in as well as kube path
	cfg, err := clientcmd.BuildConfigFromFlags(kubeStartUpConfig.masterUrl, kubeStartUpConfig.k8sConfigPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialise client from config: %s", err.Error())
	}
	if b, err := os.ReadFile(kubeStartUpConfig.k8sConfigPath); err != nil {
		logger.Errorf("kubeConfigPath err: %v", err)
	} else {
		logger.Infof("kubeConfigPath file (%s) contents: %v", kubeStartUpConfig.k8sConfigPath, string(b))
	}
	logger.Infof("custom config\n\nHost: %v\nServeName: %s", cfg.Host, cfg.ServerName)

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
	kubeStartUpConfig = DetermineKubeConfig()
	logf.SetLogger(logr.WithName("RestStrategyController-Test"))
	deleteCluster = startCluster(t)

	kubeClient, cfg, e := kubeClientSetup(t)
	if e != nil {
		t.Errorf("failed to get client: %v", e)
	}

	logger.Infof("config returned from kubeClient setup: %v", cfg)

	var keepTrying, attempts = true, 0

	for keepTrying {
		attempts++
		logger.Infof("attemp: %d", attempts)
		pod, err := kubeClient.CoreV1().Pods("kube-system").Get(context.Background(), fmt.Sprintf("kube-apiserver-%s-control-plane", defaultClusterName), v1.GetOptions{})
		if err != nil {
			logger.Error(err)
			keepTrying = false
		}
		if attempts < 10 {
			logger.Infof("status\nPhase: %s\nMessage: %s\nReason: %s\n", pod.Status.Phase, pod.Status.Message, pod.Status.Reason)
			logger.Debugf("pod: %v", pod)
			keepTrying = pod.Status.Phase != podv1.PodRunning
		} else {
			logger.Infof("status\nPhase: %s\nMessage: %s\nReason: %s\n", pod.Status.Phase, pod.Status.Message, pod.Status.Reason)
			logger.Infof("attemps depleted")
			keepTrying = false
		}
		time.Sleep(time.Duration(time.Second * 2))
	}

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
	if deleteCluster != nil {
		deleteCluster()
	}
})
