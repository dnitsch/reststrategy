package k8sutils

import (
	"errors"
	"os"
	"time"

	clientset "github.com/dnitsch/reststrategy/apis/generated/clientset/versioned"

	informers "github.com/dnitsch/reststrategy/apis/generated/informers/externalversions"
)

func InitialiseSharedInformerFactory(reststrategyClient clientset.Interface, namespace string, resyncCustom time.Duration) (informers.SharedInformerFactory, error) {
	// var opts informers.SharedInformerOption
	if ns := getNamespace(namespace); ns != "" {
		// informers.WithCustomResyncConfig()
		informers.WithNamespace(namespace)
		options := informers.WithNamespace(namespace)
		return informers.NewSharedInformerFactoryWithOptions(reststrategyClient, time.Second*resyncCustom, options), nil
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
