package cluster

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"
	"testing"
	"time"

	log "github.com/dnitsch/simplelog"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"

	podv1 "k8s.io/api/core/v1"
)

type ClusterTester struct {
	Name     string
	EnsureUp bool
	logger   log.Loggeriface
}

func (c *ClusterTester) WithLogger(logger log.Loggeriface) *ClusterTester {
	c.logger = logger
	return c
}

// ====
// BEGIN CUSTOM K8s setup
type KubeConfig struct {
	isInCI        bool
	k8sConfigPath string
	kindConfig    *v1alpha4.Cluster
	masterUrl     string
}

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

func (c *ClusterTester) DetermineKubeConfig() KubeConfig {
	usr, _ := user.Current()
	hd := usr.HomeDir
	kubeConfigPath := path.Join(hd, ".kube/config")
	resp := KubeConfig{isInCI: false, k8sConfigPath: kubeConfigPath, kindConfig: nil, masterUrl: ""}
	// if len(os.Getenv("GITHUB_ACTIONS")) > 0 ||
	// 	len(os.Getenv("TRAVIS")) > 0 ||
	// 	len(os.Getenv("CIRCLECI")) > 0 ||
	// 	len(os.Getenv("GITLAB_CI")) > 0 ||
	// 	len(os.Getenv("CI")) > 0 {
	// 	logger.Info("In CI will be using a service mounted KinD")
	// 	resp.kindConfig = &v1alpha4.Cluster{
	// 		Networking: v1alpha4.Networking{
	// 			APIServerAddress: "127.0.0.1",
	// 			APIServerPort:    6443,
	// 		},
	// 	}
	// 	resp.masterUrl = "https://kind-control-plane:6443"
	// 	return resp
	// }
	return resp
}

// start kind cluster
func (c *ClusterTester) StartCluster(t *testing.T, kubeStartUpConfig KubeConfig) func() {
	impProvider := detectContainerImp()
	if impProvider == nil {
		c.logger.Info("unable to find suitable containerisation provider")
		return func() {}
	}

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
		c.Name,
		clusterCreateOptions...,
	); err != nil {
		c.logger.Errorf("failed to create cluster: %v", err)
		t.Fatal(errors.Wrap(err, "failed to create cluster"))
	}
	return func() {
		// delete cluster
		if err := provider.Delete(c.Name, kubeStartUpConfig.k8sConfigPath); err != nil {
			t.Errorf("failed to tear down kind cluster: %s", err)
		}
	}
}

// k8s-client set up
func (c *ClusterTester) KubeClientSetup(t *testing.T, kubeStartUpConfig KubeConfig) (*kubernetes.Clientset, *rest.Config, error) {

	cfg, err := clientcmd.BuildConfigFromFlags(kubeStartUpConfig.masterUrl, kubeStartUpConfig.k8sConfigPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialise client from config: %s", err.Error())
	}
	if b, err := os.ReadFile(kubeStartUpConfig.k8sConfigPath); err != nil {
		c.logger.Errorf("kubeConfigPath err: %v", err)
	} else {
		c.logger.Infof("kubeConfigPath file (%s) contents: %v", kubeStartUpConfig.k8sConfigPath, string(b))
	}
	c.logger.Infof("custom config\n\nHost: %v\nServeName: %s", cfg.Host, cfg.ServerName)

	kubeClient, err := kubernetes.NewForConfig(cfg)

	if err != nil {
		return nil, nil, fmt.Errorf("error building kubernetes clientset: %s", err.Error())
	}

	c.logger.Infof("config kubeClient: %v", kubeClient)
	return kubeClient, cfg, nil
}

func (c *ClusterTester) EnsureReachable(kubeClient *kubernetes.Clientset) bool {
	if !c.EnsureUp {
		return true
	}

	var keepTrying, attempts = true, 0

	for keepTrying {
		attempts++
		c.logger.Debugf("attemp: %d", attempts)
		pod, err := kubeClient.CoreV1().Pods("kube-system").Get(context.Background(), fmt.Sprintf("kube-apiserver-%s-control-plane", c.Name), v1.GetOptions{})
		if err != nil {
			c.logger.Error(err)
			keepTrying = false
		}
		if attempts < 10 {
			c.logger.Debugf("status\nPhase: %s\nMessage: %s\nReason: %s\n", pod.Status.Phase, pod.Status.Message, pod.Status.Reason)
			c.logger.Debugf("pod: %v", pod)
			keepTrying = pod.Status.Phase != podv1.PodRunning
		} else {
			c.logger.Infof("status\nPhase: %s\nMessage: %s\nReason: %s\n", pod.Status.Phase, pod.Status.Message, pod.Status.Reason)
			c.logger.Infof("attemps depleted")
			keepTrying = false
		}
		time.Sleep(time.Duration(time.Second * 2))
	}
	return false

}

//
// END CUSTOM K8s setup
// ====
