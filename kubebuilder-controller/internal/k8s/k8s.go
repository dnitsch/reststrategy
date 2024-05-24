package k8s

import (
	"fmt"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/dnitsch/configmanager/pkg/generator"
	seederv1alpha1 "github.com/dnitsch/reststrategy/kubebuilder-controller/api/v1alpha1"
	"github.com/dnitsch/reststrategy/kubebuilder-controller/controllers"
	"github.com/go-logr/logr"

	log "github.com/dnitsch/simplelog"
)

const CACHE_RESYNC_INTERVAL int = 10

type Config struct {
	MasterURL            string
	Kubeconfig           string
	ControllerCount      int
	Rsyncperiod          int
	Namespace            string
	LogLevel             string
	ProbeAddr            string
	MetricsAddr          string
	EnableLeaderElection bool
	ConfigManager        *generator.GenVarsConfig
}

func Run(conf Config, logger logr.Logger, scheme *runtime.Scheme) {
	// removed init
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(seederv1alpha1.AddToScheme(scheme))

	// fs := &flag.FlagSet{}
	// ctrl.RegisterFlags(fs)

	ctrl.SetLogger(logger.WithName("RestStrategyController"))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     conf.MetricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: conf.ProbeAddr,
		LeaderElection:         conf.EnableLeaderElection,
		LeaderElectionID:       "f1b2a8fa.dnitsch.net",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
		Namespace: GetNamespace(conf.Namespace),
		// SyncPeriod is left empty by defult in this instance
		// as the specific use case of periodic resync is better handled by
		// specifying a RequeueAfter time.Duration. this allows for more
		// controlled by resource periodic resync and ensure the state on the
		// remote is periodically synced with the desired state
		// SyncPeriod: nil
	})
	if err != nil {
		ctrl.Log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.RestStrategyReconciler{
		Client:                     mgr.GetClient(),
		Scheme:                     mgr.GetScheme(),
		ResyncPeriod:               conf.Rsyncperiod,
		Logger:                     log.New(os.Stderr, log.DebugLvl),
		FailedResourceResyncPeriod: 10,
		ConfigManagerConfig:        conf.ConfigManager,
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create controller", "controller", "RestStrategy")
		MsgExit(ctrl.Log, 1, "CreateControllerError")
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		ctrl.Log.Error(err, "unable to set up health check")
		MsgExit(ctrl.Log, 1, "AddHealthzCheck")
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		ctrl.Log.Error(err, "unable to set up ready check")
		MsgExit(ctrl.Log, 1, "AddReadyzCheck")

	}

	ctrl.Log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		// setupLog.Error(err, "problem running manager")
		ctrl.Log.Error(err, "problem running manager")
		MsgExit(ctrl.Log, 1, "startingManager")
	}
}

// GetNamespace returns a namespace to use in controller manager set up
func GetNamespace(namespace string) string {
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

func MsgExit(logger logr.Logger, code int, area string) {
	if _, ok := os.LookupEnv("IN_CI_TEST"); ok {
		logger.Info(fmt.Sprintf("In CI tests, do not exit with code: %d\n", code), "code", code, "area", area)
		return
	}
	os.Exit(code)
}
