package controller

import (
	"fmt"
	"os"

	"github.com/dnitsch/reststrategy/controller/internal/k8sutils"
	"github.com/dnitsch/reststrategy/controller/pkg/signals"
	log "github.com/dnitsch/simplelog"
	"github.com/spf13/cobra"
)

var (
	masterURL       string
	kubeconfig      string
	controllerCount int
	rsyncperiod     int
	namespace       string
	logLevel        string

	controllerCmd = &cobra.Command{
		Short: fmt.Sprintf("%s starts the controller", "reststrategy-controller"),
		Long:  fmt.Sprintf(`%s CLI provides an idempotent rest caller for seeding configuration or data in a repeatable manner`, "reststrategy-controller"),
		RunE:  controllerRun,
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
	controllerCmd.PersistentFlags().IntVarP(&controllerCount, "controllercount", "c", 2, "Number of spawned go routines for the controller")
	controllerCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Only process RestStrategy in this namespace. Overrides using the controllers own namespace")
	controllerCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "l", "error", "The severity threshold for emitting log events")
	controllerCmd.PersistentFlags().IntVarP(&rsyncperiod, "rsync", "r", 12, "Period for resyncing the controller periodically even if there are no changes. Vaule is for hours and used inside a modulo computation on a 24 day :) - will not work on Mars as intended")
}

func controllerRun(cmd *cobra.Command, args []string) error {
	logger := log.New(os.Stderr, log.ParseLevel(logLevel))

	// great case for Viper here
	config := k8sutils.Config{
		Kubeconfig:      kubeconfig,
		MasterURL:       masterURL,
		ControllerCount: controllerCount,
		Rsyncperiod:     rsyncperiod,
		Namespace:       namespace,
		LogLevel:        logLevel,
	}
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	return k8sutils.Run(config, logger, stopCh)
}
