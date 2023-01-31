package k8sutils

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	clientset "github.com/dnitsch/reststrategy/apis/reststrategy/generated/clientset/versioned"
	controllerinformers "github.com/dnitsch/reststrategy/apis/reststrategy/generated/informers/externalversions"
	"github.com/dnitsch/reststrategy/controller"
	log "github.com/dnitsch/simplelog"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const CACHE_RESYNC_INTERVAL int = 10

type Config struct {
	MasterURL       string
	Kubeconfig      string
	ControllerCount int
	Rsyncperiod     int
	Namespace       string
	LogLevel        string
}

// Run accepts config object and logger implementation and stop channel 
func Run(config Config, log log.Loggeriface, stopCh <-chan struct{}) error {

	cfg, err := clientcmd.BuildConfigFromFlags(config.MasterURL, config.Kubeconfig)
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

	reststrategyInformerFactory, err := initialiseSharedInformerFactory(reststrategyClient, config.Namespace, 60)
	if err != nil {
		fmt.Print(fmt.Errorf("error building reststrategyInformerFactory: %s", err.Error()))
		os.Exit(1)
	}

	controller := controller.NewController(kubeClient, reststrategyClient,
		reststrategyInformerFactory.Reststrategy().V1alpha1().RestStrategies(), config.Rsyncperiod)

	rc := &http.Client{}
	controller.WithLogger(log).WithRestClient(rc)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	// Own Implementation with cancellation token
	reststrategyInformerFactory.Start(stopCh)

	// potentially additional go routines may need to be created
	// TODO: might need to config the number of goroutines
	return controller.Run(config.ControllerCount, stopCh)

}

func initialiseSharedInformerFactory(reststrategyClient clientset.Interface, namespace string, resyncCustom time.Duration) (controllerinformers.SharedInformerFactory, error) {

	if ns := getNamespace(namespace); ns != "" {
		options := controllerinformers.WithNamespace(namespace)
		return controllerinformers.NewSharedInformerFactoryWithOptions(reststrategyClient, time.Second*resyncCustom, options), nil
	}
	return nil, errors.New("either --namespace arg must be provided or POD_NAMESPACE env variable must be present")
}

func getNamespace(namespace string) string {
	ns := ""
	if len(namespace) > 0 {
		ns = namespace
	}
	if len(ns) < 1 {
		if podNamespace, exists := os.LookupEnv("POD_NAMESPACE"); exists {
			ns = podNamespace
		}
	}
	return ns
}
