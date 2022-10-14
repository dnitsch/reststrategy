/*
influenced by k8s.io samplecontroller
*/

package controller

import (
	"fmt"
	"time"

	"github.com/dnitsch/configmanager"
	"github.com/dnitsch/configmanager/pkg/generator"
	clientset "github.com/dnitsch/reststrategy/apis/reststrategy/generated/clientset/versioned"
	reststrategyscheme "github.com/dnitsch/reststrategy/apis/reststrategy/generated/clientset/versioned/scheme"
	informers "github.com/dnitsch/reststrategy/apis/reststrategy/generated/informers/externalversions/reststrategy/v1alpha1"
	listers "github.com/dnitsch/reststrategy/apis/reststrategy/generated/listers/reststrategy/v1alpha1"
	"github.com/dnitsch/reststrategy/apis/reststrategy/v1alpha1"
	"github.com/dnitsch/reststrategy/controller/pkg/rstservice"
	"github.com/dnitsch/reststrategy/seeder/pkg/rest"

	log "github.com/dnitsch/simplelog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "reststrategycontroller"
const kindCrdName = "RestStrategies"

var (
	Version  string = "0.0.1"
	Revision string = "1111aaaa"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a RestStrategy is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a RestStrategy fails
	// to sync due to a Deployment of the same name already existing.
	ErrSync = "Failed"

	// FailedToSyncAll is the message used for Events when a resource
	// fails to sync due to not all resources syncing correctly
	FailedToSyncAll = "Failed due to not all subcomponents within RestStrategy successfully completed. see additionalInfo for more details"
	FailReason      = "SubcomponentsFailed"
	// MessageResourceSynced is the message used for an Event fired when a RestStrategy
	// is synced successfully
	MessageResourceSynced = "RestStrategy successfully executed"
	MessageResourceFailed = "RestStrategy failed to sync"
)

// Controller is the controller implementation for RestStrategy resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// reststrategyclientset is a clientset for our own API group
	reststrategyclientset clientset.Interface
	// Lister for own resources
	reststrategysLister listers.RestStrategyLister
	reststrategysSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
	// logger - init with Standard
	log log.Logger
	// service struct that will perform the business logic
	restClient rest.Client
	// resyncPeriod in hours
	resyncServicePeriod int
}

// NewController returns a new RestStrategy controller
func NewController(
	kubeclientset kubernetes.Interface,
	reststrategyclientset clientset.Interface,
	// own CRD informer
	reststrategyInformer informers.RestStrategyInformer,
	resyncServicePeriodHours int,
) *Controller {

	// Create event broadcaster
	// Add reststrategy-controller types to the default Kubernetes Scheme so Events can be
	// logged for reststrategy-controller types.
	utilruntime.Must(reststrategyscheme.AddToScheme(scheme.Scheme))
	fmt.Print("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:         kubeclientset,
		reststrategyclientset: reststrategyclientset,
		reststrategysLister:   reststrategyInformer.Lister(),
		reststrategysSynced:   reststrategyInformer.Informer().HasSynced,
		workqueue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), kindCrdName),
		recorder:              recorder,
		resyncServicePeriod:   resyncServicePeriodHours,
	}

	fmt.Print("Setting up event handlers")
	// Set up an event handler for when RestStrategy resources change
	reststrategyInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.createCustomResource,
		UpdateFunc: controller.updateCustomResource,
		DeleteFunc: controller.deleteRestStrategyCustomResource,
	})
	return controller
}

// WithLogger overwrites logger with Custom implementation
func (c *Controller) WithLogger(l log.Logger) *Controller {
	c.log = l
	return c
}

// WithService assigns a service instance
func (c *Controller) WithRestClient(rc rest.Client) *Controller {
	c.restClient = rc
	return c
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	c.log.Info("Starting RestStrategy controller")

	// Wait for the caches to be synced before starting workers
	c.log.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.reststrategysSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	c.log.Info("Starting workers")
	// Launch X workers to process RestStrategy resources
	//
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	c.log.Info("Started workers")
	<-stopCh
	c.log.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	c.log.Debug("processing items")
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// RestStrategy resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			// NOTE: it might make more sense to break the
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%v'\n\nequeuing: %v", key, err)
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		c.log.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the RestStrategy resource
// with the current status of the resource.
// this is where the meat of the logic happens ==> need to hand over to another pkg which has the relevant business domain knowledge
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the RestStrategy resource with this namespace/name
	reststrategy, err := c.reststrategysLister.RestStrategies(namespace).Get(name)
	if err != nil {
		// The RestStrategy resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("reststrategy '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	// THIS IS WHERE WE need to hand over to the domain aware service
	// create Copy to pass to service
	reststrategyCopy := reststrategy.DeepCopy()

	c.log.Debugf("Handing over resource: '%s' in namespace: '%s' to the service handler. allocating new srv instance", name, namespace)

	rstsrv := rstservice.New(c.log, c.restClient)

	cm := &configmanager.ConfigManager{}

	// use custom token separator inline with future releases
	config := generator.NewConfig().WithTokenSeparator("://")

	rspec, err := configmanager.KubeControllerSpecHelper(reststrategyCopy.Spec, cm, *config)

	if err != nil {
		c.log.Debugf("failed to replace any found tokens on the CRD spec: %s", key)
		return err
	}

	if err := rstsrv.Execute(*rspec); err != nil {
		c.log.Errorf("%+#v", err)
		c.recorder.Event(reststrategy, corev1.EventTypeNormal, ErrSync, fmt.Sprintf("#+%v", err))
	} else {
		c.recorder.Event(reststrategy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	}

	// force quicker GC ... deallocation from heap
	rstsrv = nil

	return nil
}

// enqueue takes a RestStrategy resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than RestStrategy.
func (c *Controller) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// createCustomResource ensure object is typecast into correct type
// and enqueued for processing
func (c *Controller) createCustomResource(new interface{}) {
	if obj, converted := new.(*v1alpha1.RestStrategy); converted {
		c.log.Infof("either booting up or new customresource is created --> %s", obj.ObjectMeta.GetName())
		c.enqueue(obj)
		return
	}
	c.log.Info("Unable to convert `obj` to v1alpha1.RestStrategy")
	c.log.Info("NotAddingToQueue")
}

// updateCustomResource does type cast
// hands over to a business aware checker
func (c *Controller) updateCustomResource(old, new interface{}) {
	oldT, newT := &v1alpha1.RestStrategy{}, &v1alpha1.RestStrategy{}
	if ot, converted := old.(*v1alpha1.RestStrategy); converted {
		oldT = ot
	}
	if nt, converted := new.(*v1alpha1.RestStrategy); converted {
		newT = nt
	}

	if c.doPeriodResync() || versionsDiffer(newT, oldT) {
		c.enqueue(new)
	}
	c.log.Debug("version of resources are the same doing nothing ...")
}

func (c *Controller) deleteRestStrategyCustomResource(obj interface{}) {
	c.log.Info("DeleteFunc -> NotImplemented")
}

// doPeriodResync ensures that a resync on target is triggered
// as per the period set. period in hours and on the full hour
// this leaves something to be desired for - it's a bit crude :)
// but with 3rd party targets for controllers, a periodic resync
// needs to happen to ensure we are all up to date and
// correct any manual settings.
func (c *Controller) doPeriodResync() bool {
	t := time.Now()
	return t.Hour()%c.resyncServicePeriod == 0 && t.Minute() == 0
}

// versionsDiffer returns true if new version is not in sync
func versionsDiffer(newDef, oldDef *v1alpha1.RestStrategy) bool {
	// check if there is a new version
	// be wary of CRDs applied via the SDK/pure API
	// may not correctly reflect Generation and ResourceVersion fields
	return oldDef.ObjectMeta.ResourceVersion != newDef.ObjectMeta.ResourceVersion && oldDef.ObjectMeta.Generation != newDef.ObjectMeta.Generation
}
