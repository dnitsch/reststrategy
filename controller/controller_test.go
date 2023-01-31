package controller

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/reststrategy/controller/internal/testutils"
	"github.com/dnitsch/reststrategy/seeder/pkg/rest"
	log "github.com/dnitsch/simplelog"

	"github.com/dnitsch/reststrategy/apis/reststrategy/generated/clientset/versioned/fake"
	informers "github.com/dnitsch/reststrategy/apis/reststrategy/generated/informers/externalversions"
	v1alphacontroller "github.com/dnitsch/reststrategy/apis/reststrategy/v1alpha1"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
	kind               = "RestStrategy"
	resource           = "reststrategies"
	group              = "reststrategy.dnitsch.net"
	version            = "v1alpha1"
)

type fixture struct {
	t *testing.T

	client     *fake.Clientset
	kubeclient *k8sfake.Clientset
	// Objects to put in the store.
	testLister []*v1alphacontroller.RestStrategy
	// Actions expected to happen on the client.
	kubeactions []core.Action
	actions     []core.Action
	// Objects from here preloaded into NewSimpleFake.
	kubeobjects []runtime.Object
	objects     []runtime.Object
	configmgr   ControllerConfigManager
}

type testFuncs struct {
	path   string
	tfuncs func(w http.ResponseWriter, r *http.Request)
}

func setupServer(t *testing.T, tf []testFuncs) http.Handler {
	mux := http.NewServeMux()

	for _, v := range tf {
		mux.HandleFunc(v.path, v.tfuncs)
	}

	return mux
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	f.kubeobjects = []runtime.Object{}
	return f
}

func newRestStrategySuccess(name, url string) *v1alphacontroller.RestStrategy {
	testAuthBasic := rest.AuthConfig{
		AuthStrategy: rest.Basic,
		Username:     "foo",
		Password:     "bar",
	}
	testSeedBasic := rest.Action{
		Strategy:          "PUT",
		Endpoint:          url,
		GetEndpointSuffix: rest.String("/get"),
		PutEndpointSuffix: rest.String("/put"),
		AuthMapRef:        "test1",
		HttpHeaders:       &map[string]string{},
		RuntimeVars:       &map[string]string{},
		Variables:         rest.KvMapVarsAny{},
	}
	return &v1alphacontroller.RestStrategy{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alphacontroller.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1alphacontroller.StrategySpec{
			AuthConfig: []v1alphacontroller.AuthConfig{
				{"test1", testAuthBasic},
			},
			Seeders: []v1alphacontroller.SeederConfig{
				{"test1-action", testSeedBasic},
				{"test2-action", testSeedBasic}},
		},
		Status: v1alphacontroller.StrategyStatus{Message: "RestStrategy successfully executed"},
	}
}

type fakeConfMgr func(input string, config generator.GenVarsConfig) (string, error)

func (f fakeConfMgr) RetrieveWithInputReplaced(input string, config generator.GenVarsConfig) (string, error) {
	return f(input, config)
}

func (f *fixture) newController() (*Controller, informers.SharedInformerFactory, kubeinformers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)

	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())
	k8sI := kubeinformers.NewSharedInformerFactory(f.kubeclient, noResyncPeriodFunc())

	c := NewController(f.kubeclient, f.client,
		i.Reststrategy().V1alpha1().RestStrategies(), 8)
	c.WithLogger(log.New(&bytes.Buffer{}, log.DebugLvl)).WithRestClient(&http.Client{})

	//&tClient{})

	c.reststrategysSynced = alwaysReady

	c.recorder = &record.FakeRecorder{}

	// scheme := runtime.NewScheme()
	// _ = v1alphacontroller.AddToScheme(scheme)

	for _, f := range f.testLister {
		i.Reststrategy().V1alpha1().RestStrategies().Informer().GetIndexer().Add(f)
	}
	// conf := generator.NewConfig().WithKeySeparator("://")
	// c.WithConfigManager(ControllerConfigManager{retrieve: fakeConfMgr(func(input string, config generator.GenVarsConfig) (string, error) {
	// 	return `{}`, nil
	// }), config: *conf})

	return c, i, k8sI
}

func (f *fixture) run(name string) {
	f.runController(name, true, false)
}

func (f *fixture) runExpectError(name string) {
	f.runController(name, true, true)
}

func (f *fixture) runController(fooName string, startInformers bool, expectError bool) {
	c, i, k8sI := f.newController()

	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
		k8sI.Start(stopCh)
	}

	err := c.syncHandler(fooName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing %s: %v", kind, err)
	} else if expectError && err == nil {
		f.t.Errorf("expected error syncing %s, got nil", kind)
	}

	actions := filterInformerActions(f.client.Actions())
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
			break
		}

		expectedAction := f.actions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}

	k8sActions := filterInformerActions(f.kubeclient.Actions())
	for i, action := range k8sActions {
		if len(f.kubeactions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(k8sActions)-len(f.kubeactions), k8sActions[i:])
			break
		}

		expectedAction := f.kubeactions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.kubeactions) > len(k8sActions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.kubeactions)-len(k8sActions), f.kubeactions[len(k8sActions):])
	}
}

// checkAction verifies that expected and actual actions are equal and both have
// same attached resources
func checkAction(expected, actual core.Action, t *testing.T) {
	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
		t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expected, actual)
		return
	}

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Errorf("Action has wrong type. Expected: %t. Got: %t", expected, actual)
		return
	}

	switch a := actual.(type) {
	case core.ListActionImpl:
		e, _ := expected.(core.ListActionImpl)
		expObject := e.GetResource()
		object := a.GetResource()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.CreateActionImpl:
		e, _ := expected.(core.CreateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.UpdateActionImpl:
		e, _ := expected.(core.UpdateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.PatchActionImpl:
		e, _ := expected.(core.PatchActionImpl)
		expPatch := e.GetPatch()
		patch := a.GetPatch()

		if !reflect.DeepEqual(expPatch, patch) {
			t.Errorf("Action %s %s has wrong patch\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expPatch, patch))
		}
	default:
		t.Errorf("Uncaptured Action %s %s, you should explicitly add a case to capture it",
			actual.GetVerb(), actual.GetResource().Resource)
	}
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", resource) ||
				action.Matches("watch", resource) ||
				action.Matches("list", kind)) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectUpdateStatusAction(rst *v1alphacontroller.RestStrategy) {
	// f.kubeactions = append(f.kubeactions, core.NewCreateAction(
	// 	schema.GroupVersionResource{Resource: "crd"},
	// 	rst.Namespace,
	// 	rst))
	// listAction := core.NewListAction(
	// 	schema.GroupVersionResource{Resource: resource, Group: "dnitsch.net", Version: version},
	// 	schema.GroupVersionKind{Kind: kind}, rst.Namespace, metav1.ListOptions{})
	updateAction := core.NewUpdateSubresourceAction(schema.GroupVersionResource{Resource: resource}, "status", rst.Namespace, rst)
	f.actions = append(f.actions, []core.Action{updateAction}...)
}

func getKey(rst *v1alphacontroller.RestStrategy, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(rst)
	if err != nil {
		t.Errorf("Unexpected error getting key for %s %v: %v", kind, rst.Name, err)
		return ""
	}
	return key
}

func TestCreatesCrdRestStrategy(t *testing.T) {

	ts := httptest.NewServer(setupServer(t, []testFuncs{{"/put", func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("Authorization") == "" {
			t.Errorf(testutils.TestPhraseWContext, "basic auth", r.Header.Get("Authorization"), "not empty")
		}
		if r.Method != "PUT" {
			t.Errorf(testutils.TestPhraseWContext, "method incorrect", r.Method, "get")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":3,"name":"fubar","a":"b","c":"d"}`))
	}}}))
	defer ts.Close()

	f := newFixture(t)
	reststrategy := newRestStrategySuccess("test", ts.URL)
	f.testLister = append(f.testLister, reststrategy)
	f.objects = append(f.objects, reststrategy)

	f.expectUpdateStatusAction(reststrategy)

	f.run(getKey(reststrategy, t))
}

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
