package controller

import (
	"fmt"
	"os"

	"github.com/dnitsch/reststrategy/kubebuilder-controller/internal/k8s"
	log "github.com/dnitsch/simplelog"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	masterURL            string
	kubeconfig           string
	rsyncperiod          int
	namespace            string
	logLevel             string
	metricsAddr          string
	enableLeaderElection bool
	probeAddr            string

	controllerCmd = &cobra.Command{
		Short: fmt.Sprintf("%s starts the controller", "reststrategy-controller"),
		Long:  fmt.Sprintf(`%s CLI provides an idempotent rest caller for seeding configuration or data in a repeatable manner`, "reststrategy-controller"),
		Run:   controllerRun,
	}
)

func Execute() {
	if err := controllerCmd.Execute(); err != nil {
		fmt.Errorf("cli error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	controllerCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	controllerCmd.PersistentFlags().StringVarP(&masterURL, "master", "m", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	controllerCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Only process RestStrategy in this namespace. Overrides using the controllers own namespace")
	controllerCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "l", "error", "The severity threshold for emitting log events")
	controllerCmd.PersistentFlags().IntVarP(&rsyncperiod, "rsync", "r", 12, "Period for resyncing the resources periodically even if there are no changes. Value is in minutes.")
	controllerCmd.PersistentFlags().StringVarP(&metricsAddr, "metrics-bind-address", "x", ":8080", "The address the metric endpoint binds to.")
	controllerCmd.PersistentFlags().StringVarP(&probeAddr, "health-probe-bind-address", "y", ":8081", "The address the probe endpoint binds to.")
	controllerCmd.PersistentFlags().BoolVarP(&enableLeaderElection, "leader-elect", "z", false, "Enable leader election for controller manager. "+
		"Enabling this will ensure there is only one active controller manager.")
}

func controllerRun(cmd *cobra.Command, args []string) {

	logger := log.NewLogr(os.Stdout, log.ParseLevel(logLevel))
	// great case for Viper here
	config := k8s.Config{
		Kubeconfig: kubeconfig,
		MasterURL:  masterURL,
		// ControllerCount: controllerCount,
		Rsyncperiod: rsyncperiod,
		Namespace:   namespace,
		LogLevel:    logLevel,
		ProbeAddr:   probeAddr,
	}
	k8s.Run(config, logger, runtime.NewScheme())
}
