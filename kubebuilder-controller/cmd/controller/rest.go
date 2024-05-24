package controller

import (
	"fmt"
	"os"

	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/reststrategy/kubebuilder-controller/internal/k8s"
	log "github.com/dnitsch/simplelog"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	Version                     string = "0.0.1"
	Revision                    string = "1111aaaa"
	masterURL                   string
	kubeconfig                  string
	rsyncperiod                 int
	namespace                   string
	logLevel                    string
	metricsAddr                 string
	enableLeaderElection        bool
	enableConfigManager         bool
	configManagerTokenSeparator string
	configManagerKeySeparator   string
	probeAddr                   string

	ControllerCmd = &cobra.Command{
		Short:   fmt.Sprintf("%s starts the controller", "reststrategy-controller"),
		Long:    fmt.Sprintf(`%s CLI provides an idempotent rest caller for seeding configuration or data in a repeatable manner`, "reststrategy-controller"),
		Run:     controllerRun,
		Version: fmt.Sprintf("%s-%s", Version, Revision),
	}
)

func Execute() {
	if err := ControllerCmd.Execute(); err != nil {
		fmt.Errorf("cli error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	ControllerCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	ControllerCmd.PersistentFlags().StringVarP(&masterURL, "master", "m", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	ControllerCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Only process RestStrategy in this namespace. Overrides using the controllers own namespace")
	ControllerCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "l", "error", "The severity threshold for emitting log events")
	ControllerCmd.PersistentFlags().IntVarP(&rsyncperiod, "rsync", "r", 12, "Period for resyncing the resources periodically even if there are no changes. Value is in minutes.")
	ControllerCmd.PersistentFlags().StringVarP(&metricsAddr, "metrics-bind-address", "x", ":8080", "The address the metric endpoint binds to.")
	ControllerCmd.PersistentFlags().StringVarP(&probeAddr, "health-probe-bind-address", "y", ":8081", "The address the probe endpoint binds to.")
	ControllerCmd.PersistentFlags().BoolVarP(&enableLeaderElection, "leader-elect", "z", false, "Enable leader election for controller manager. "+
		"Enabling this will ensure there is only one active controller manager.")
	ControllerCmd.PersistentFlags().BoolVarP(&enableConfigManager, "configmanager", "c", false, "Enable configmanager on resources handling via the restseeder.")
	ControllerCmd.PersistentFlags().StringVarP(&configManagerKeySeparator, "configmanager-key-separator", "", "|", "If configmanager is enabled - this key separator will be used.")
	ControllerCmd.PersistentFlags().StringVarP(&configManagerTokenSeparator, "configmanager-token-separator", "t", "://", "If configmanager is enabled - this key separator will be used.")
}

func controllerRun(cmd *cobra.Command, args []string) {

	logger := log.NewLogr(os.Stdout, log.ParseLevel(logLevel))

	// if kubeconfig is set
	// we want to overwrite the ENV variable with this value
	if kubeconfig != "" {
		os.Setenv("KUBECONFIG", kubeconfig)
	}
	// great case for Viper here
	config := k8s.Config{
		Kubeconfig: kubeconfig,
		MasterURL:  masterURL,
		// ControllerCount: controllerCount,
		Rsyncperiod:   rsyncperiod,
		Namespace:     namespace,
		LogLevel:      logLevel,
		ProbeAddr:     probeAddr,
		ConfigManager: nil,
	}
	if enableConfigManager {
		config.ConfigManager = generator.NewConfig().WithKeySeparator(configManagerKeySeparator).WithTokenSeparator(configManagerTokenSeparator)
	}
	k8s.Run(config, logger, runtime.NewScheme())
}
