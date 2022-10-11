/*
influenced by k8s.io samplecontroller
*/

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	clientset "github.com/dnitsch/reststrategy/apis/generated/clientset/versioned"
	"github.com/dnitsch/reststrategy/controller/internal/k8sutils"
	"github.com/dnitsch/reststrategy/controller/pkg/signals"
	log "github.com/dnitsch/simplelog"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const CACHE_RESYNC_INTERVAL int = 10

var (
	masterURL       string
	kubeconfig      string
	controllerCount int
	rsyncperiod     int
	namespace       string
	logLevel        string
)

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		fmt.Print(fmt.Errorf("error building kubeconfig: %s", err.Error()))
		os.Exit(1)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Print(fmt.Errorf("error building kubernetes clientset: %s", err.Error()))
		os.Exit(1)
	}

	reststrategyClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		fmt.Print(fmt.Errorf("error building clientset: %s", err.Error()))
		os.Exit(1)
	}

	// set cache resync period to 60 seconds so that the resyncPeriod
	// can be more easily satisfied without the need to keep a record
	// current runs etc..
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Minute*time.Duration(CACHE_RESYNC_INTERVAL))

	reststrategyInformerFactory, err := k8sutils.InitialiseSharedInformerFactory(reststrategyClient, namespace, 60)
	if err != nil {
		fmt.Print(fmt.Errorf("error building reststrategyInformerFactory: %s", err.Error()))
		os.Exit(1)
	}

	controller := NewController(kubeClient, reststrategyClient,
		reststrategyInformerFactory.Reststrategy().V1alpha1().RestStrategies(), rsyncperiod)

	logger := log.New(os.Stderr, log.ParseLevel(logLevel))
	rc := &http.Client{}
	controller.WithLogger(logger).WithService(rc)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	// Own Implementation with cancellation token
	reststrategyInformerFactory.Start(stopCh)

	// potentially additional go routines may need to be created
	// TODO: might need to config the number of goroutines
	if err = controller.Run(controllerCount, stopCh); err != nil {
		fmt.Print(fmt.Errorf("error running controller: %v", err.Error()))
		os.Exit(1)
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.IntVar(&controllerCount, "controllercount", 2, "Number of spawned go routines for the controller")
	flag.StringVar(&namespace, "namespace", "", "Only process RestStrategy in this namespace. Overrides using the controllers own namespace")
	flag.StringVar(&logLevel, "loglevel", "error", "The severity threshold for emitting log events")
	flag.IntVar(&rsyncperiod, "rsync", 12, "Period for resyncing the controller periodically even if there are no changes. Vaule is for hours and used inside a modulo computation on a 24 day :) - will not work on Mars as intended")
}
