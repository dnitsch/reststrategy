package main

// import (

// 	"reflect"
// 	"testing"
// 	"time"

// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/apimachinery/pkg/runtime/schema"
// 	"k8s.io/apimachinery/pkg/util/diff"
// 	kubeinformers "k8s.io/client-go/informers"
// 	k8sfake "k8s.io/client-go/kubernetes/fake"
// 	core "k8s.io/client-go/testing"
// 	"k8s.io/client-go/tools/cache"
// 	"k8s.io/client-go/tools/record"

// 	"github.com/dnitsch/reststrategy/controller/internal/testutils"

// 	"github.com/dnitsch/reststrategy/apis/generated/clientset/versioned/fake"
// 	informers "github.com/dnitsch/reststrategy/apis/generated/informers/externalversions"
// 	v1alphacontroller "github.com/dnitsch/reststrategy/apis/reststrategy/v1alpha1"
// )

// var (
// 	alwaysReady        = func() bool { return true }
// 	noResyncPeriodFunc = func() time.Duration { return 0 }
// 	resourceKind       = "RestStrategy"
// )

// type fixture struct {
// 	t *testing.T

// 	client     *fake.Clientset
// 	kubeclient *k8sfake.Clientset
// 	// Objects to put in the store.
// 	testLister []*v1alphacontroller.RestStrategy
// 	// Actions expected to happen on the client.
// 	kubeactions []core.Action
// 	actions     []core.Action
// 	// Objects from here preloaded into NewSimpleFake.
// 	kubeobjects []runtime.Object
// 	objects     []runtime.Object
// 	orcConfig   *config.Config
// }

// func newFixture(t *testing.T) *fixture {
// 	f := &fixture{}
// 	f.t = t
// 	f.objects = []runtime.Object{}
// 	f.kubeobjects = []runtime.Object{}
// 	return f
// }

// func newRestStrategySuccess(name string) *v1alphacontroller.RestStrategy {
// 	onboardappSpec := &v1alphacontroller.RestStrategySpec{}
// 	file, err := testutils.TestGoodInputFile()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	input, err := testutils.AppOnboardJson(file, onboardappSpec)
// 	if err != nil {
// 		log.Fatalf("Err: %s", err)
// 	}

// 	return &v1alphacontroller.RestStrategy{
// 		TypeMeta: metav1.TypeMeta{APIVersion: v1alphacontroller.SchemeGroupVersion.String()},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      name,
// 			Namespace: metav1.NamespaceDefault,
// 		},
// 		Spec: *input,
// 	}
// }

// func (f *fixture) newController() (*Controller, informers.SharedInformerFactory, kubeinformers.SharedInformerFactory) {
// 	f.client = fake.NewSimpleClientset(f.objects...)
// 	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)

// 	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())
// 	k8sI := kubeinformers.NewSharedInformerFactory(f.kubeclient, noResyncPeriodFunc())

// 	c := NewController(f.kubeclient, f.client,
// 		i.reststrategy().V1alpha1().RestStrategys())

// 	c.reststrategysSynced = alwaysReady

// 	c.recorder = &record.FakeRecorder{}

// 	for _, f := range f.testLister {
// 		i.reststrategy().V1alpha1().RestStrategys().Informer().GetIndexer().Add(f)
// 	}

// 	return c, i, k8sI
// }

// func (f *fixture) run(name string) {
// 	f.runController(name, true, false)
// }

// func (f *fixture) runExpectError(name string) {
// 	f.runController(name, true, true)
// }

// func (f *fixture) runController(fooName string, startInformers bool, expectError bool) {
// 	c, i, k8sI := f.newController()

// 	srv := testutils.SetHttpUpMockServer(testutils.SetUpHandlerFuncPositiveResponse(`{"message": "oapp", "success": true}`))
// 	defer srv.Close()

// 	f.orcConfig = config.New(srv.URL, 12)

// 	c.WithOrchestrator(*f.orcConfig)

// 	// c.WithOrchestrator(*f.orcConfig)
// 	if startInformers {
// 		stopCh := make(chan struct{})
// 		defer close(stopCh)
// 		i.Start(stopCh)
// 		k8sI.Start(stopCh)
// 	}

// 	err := c.syncHandler(fooName)
// 	if !expectError && err != nil {
// 		f.t.Errorf("error syncing %s: %v", resourceKind, err)
// 	} else if expectError && err == nil {
// 		f.t.Errorf("expected error syncing %s, got nil", resourceKind)
// 	}

// 	actions := filterInformerActions(f.client.Actions())
// 	for i, action := range actions {
// 		if len(f.actions) < i+1 {
// 			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
// 			break
// 		}

// 		expectedAction := f.actions[i]
// 		checkAction(expectedAction, action, f.t)
// 	}

// 	if len(f.actions) > len(actions) {
// 		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
// 	}

// 	k8sActions := filterInformerActions(f.kubeclient.Actions())
// 	for i, action := range k8sActions {
// 		if len(f.kubeactions) < i+1 {
// 			f.t.Errorf("%d unexpected actions: %+v", len(k8sActions)-len(f.kubeactions), k8sActions[i:])
// 			break
// 		}

// 		expectedAction := f.kubeactions[i]
// 		checkAction(expectedAction, action, f.t)
// 	}

// 	if len(f.kubeactions) > len(k8sActions) {
// 		f.t.Errorf("%d additional expected actions:%+v", len(f.kubeactions)-len(k8sActions), f.kubeactions[len(k8sActions):])
// 	}
// }

// // checkAction verifies that expected and actual actions are equal and both have
// // same attached resources
// func checkAction(expected, actual core.Action, t *testing.T) {
// 	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
// 		t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expected, actual)
// 		return
// 	}

// 	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
// 		t.Errorf("Action has wrong type. Expected: %t. Got: %t", expected, actual)
// 		return
// 	}

// 	switch a := actual.(type) {
// 	case core.CreateActionImpl:
// 		e, _ := expected.(core.CreateActionImpl)
// 		expObject := e.GetObject()
// 		object := a.GetObject()

// 		if !reflect.DeepEqual(expObject, object) {
// 			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
// 				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
// 		}
// 	case core.UpdateActionImpl:
// 		e, _ := expected.(core.UpdateActionImpl)
// 		expObject := e.GetObject()
// 		object := a.GetObject()

// 		if !reflect.DeepEqual(expObject, object) {
// 			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
// 				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
// 		}
// 	case core.PatchActionImpl:
// 		e, _ := expected.(core.PatchActionImpl)
// 		expPatch := e.GetPatch()
// 		patch := a.GetPatch()

// 		if !reflect.DeepEqual(expPatch, patch) {
// 			t.Errorf("Action %s %s has wrong patch\nDiff:\n %s",
// 				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expPatch, patch))
// 		}
// 	default:
// 		t.Errorf("Uncaptured Action %s %s, you should explicitly add a case to capture it",
// 			actual.GetVerb(), actual.GetResource().Resource)
// 	}
// }

// // filterInformerActions filters list and watch actions for testing resources.
// // Since list and watch don't change resource state we can filter it to lower
// // nose level in our tests.
// func filterInformerActions(actions []core.Action) []core.Action {
// 	ret := []core.Action{}
// 	for _, action := range actions {
// 		if len(action.GetNamespace()) == 0 &&
// 			(action.Matches("list", resourceKind) ||
// 				action.Matches("watch", resourceKind)) {
// 			continue
// 		}
// 		ret = append(ret, action)
// 	}

// 	return ret
// }

// func (f *fixture) expectUpdateOAappStatusAction(oapp *v1alphacontroller.RestStrategy) {
// 	f.kubeactions = append(f.kubeactions, core.NewCreateAction(schema.GroupVersionResource{Resource: "crd"}, oapp.Namespace, oapp))
// 	action := core.NewUpdateSubresourceAction(schema.GroupVersionResource{Resource: "onboardapps"}, "status", oapp.Namespace, oapp)
// 	f.actions = append(f.actions, action)
// }

// func getKey(oapp *v1alphacontroller.RestStrategy, t *testing.T) string {
// 	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(oapp)
// 	if err != nil {
// 		t.Errorf("Unexpected error getting key for %s %v: %v", resourceKind, oapp.Name, err)
// 		return ""
// 	}
// 	return key
// }

// // Tests need fixing up a bit
// // TODO: enable these top level tests when ready
// func TestCreatesCrdRestStrategy(t *testing.T) {
// 	f := newFixture(t)
// 	oapp := newRestStrategySuccess("test")

// 	f.testLister = append(f.testLister, oapp)
// 	f.objects = append(f.objects, oapp)

// 	f.expectUpdateOAappStatusAction(oapp)

// 	// f.run(getKey(oapp, t))
// 	if false {
// 		t.Errorf("Skipped tests for now")
// 	}
// }

// func TestDoNothing(t *testing.T) {
// 	f := newFixture(t)
// 	oapp := newRestStrategySuccess("app-abc")

// 	f.testLister = append(f.testLister, oapp)
// 	f.objects = append(f.objects, oapp)
// 	// append(f.kubeclient.Resources, )
// 	f.expectUpdateOAappStatusAction(oapp)
// 	// f.run(getKey(oapp, t))
// 	if false {
// 		t.Errorf("Skipped tests for now")
// 	}
// }

// func TestNotControlledByUs(t *testing.T) {
// 	f := newFixture(t)
// 	oapp := newRestStrategySuccess("test")

// 	f.testLister = append(f.testLister, oapp)
// 	f.objects = append(f.objects, oapp)

// 	// f.runExpectError(getKey(oapp, t))
// 	if false {
// 		t.Errorf("Skipped tests for now")
// 	}
// }
