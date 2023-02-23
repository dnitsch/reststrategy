package controllers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	seederv1alpha1 "github.com/dnitsch/reststrategy/kubebuilder-controller/api/v1alpha1"
	"github.com/dnitsch/reststrategy/seeder"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Define utility constants for object names and testing timeouts/durations and intervals.
const (
	RestStrategyName      = "test-rest-strategy-mocked"
	RestStrategyNameFail  = "test-rest-strategy-mocked-fail"
	RestStrategyNamespace = "default"

	timeout  = time.Second * 10
	duration = time.Second * 10
	interval = time.Millisecond * 250
)

func succesEndpoint() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`[{"foo":"bar","id":1234}]`))
	})
	mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"foo":"bar","id":1234}`))
	})
	mux.HandleFunc("/patch/1234", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"foo":"bar","id":1234}`))
	})
	return mux
}

var _ = Describe("RestStrategy controller", func() {

	Context("When updating RestStrategy Status", func() {
		It("Should correctly pass the rest seeding", func() {
			By("By creating a new RestStrategy")
			ctx := context.Background()
			crdSpec := &seederv1alpha1.RestStrategy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "seeder.dnitsch.net/v1alpha",
					Kind:       "RestStrategy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      RestStrategyName,
					Namespace: RestStrategyNamespace,
				},
				Spec: seederv1alpha1.StrategySpec{
					AuthConfig: []seeder.AuthConfig{
						{
							Name:         "basic",
							AuthStrategy: seeder.Basic,
							Username:     "foo",
							Password:     "bar",
						},
					},
					Seeders: []seeder.Action{},
				},
			}
			err := k8sClient.Create(ctx, crdSpec)
			Expect(err).Should(BeNil())

			specLookupKey := types.NamespacedName{Name: RestStrategyName, Namespace: RestStrategyNamespace}
			createdSpec := &seederv1alpha1.RestStrategy{}

			// We'll need to retry getting this newly created RestStrategy, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, specLookupKey, createdSpec)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdSpec.Spec.AuthConfig[0].Name).Should(Equal("basic"))
			By("By checking the UpdatedRestStrategy")
			Consistently(func() (string, error) {
				updatedSpec := &seederv1alpha1.RestStrategy{}
				err := k8sClient.Get(ctx, specLookupKey, updatedSpec)
				if err != nil {
					return "", err
				}
				return updatedSpec.Status.Message, nil
			}, duration, interval).Should(Equal(fmt.Sprintf(SuccessMessage, RestStrategyName, RestStrategyNamespace)))
		})
	})

	Context("When supplying incorrect runtime RestStrategy spec", func() {
		It("Should fail the rest seeding", func() {
			By("By creating a new RestStrategy")
			ctx := context.Background()
			ts := httptest.NewServer(succesEndpoint())
			defer ts.Close()
			crdSpec := &seederv1alpha1.RestStrategy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "seeder.dnitsch.net/v1alpha",
					Kind:       "RestStrategy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      RestStrategyNameFail,
					Namespace: RestStrategyNamespace,
				},
				Spec: seederv1alpha1.StrategySpec{
					AuthConfig: []seeder.AuthConfig{
						{
							Name:         "basic",
							AuthStrategy: seeder.Basic,
							Username:     "foo",
							Password:     "bar",
						},
					},
					Seeders: []seeder.Action{
						{
							Name:                 "test_pass",
							Endpoint:             ts.URL,
							Strategy:             "FIND/PATCH/POST",
							GetEndpointSuffix:    seeder.String("/get"),
							FindByJsonPathExpr:   "$.items[?(@.foo=='bar')].id",
							AuthMapRef:           "basic",
							PayloadTemplate:      `{"foo": "baz","id":1234}`,
							PatchPayloadTemplate: `{"foo": "baz","id":1234}`,
							Variables:            seeder.KvMapVarsAny{},
							RuntimeVars:          map[string]string{},
						},
						{
							Name:                 "test_pass",
							Endpoint:             ts.URL,
							Strategy:             "FIND/PATCH/POST",
							GetEndpointSuffix:    seeder.String("/get"),
							FindByJsonPathExpr:   "$.items[?(@.foo=='bar')].id",
							AuthMapRef:           "wrong_auth",
							PayloadTemplate:      `{"foo": "baz","id":1234}`,
							PatchPayloadTemplate: `{"foo": "baz","id":1234}`,
							Variables:            seeder.KvMapVarsAny{},
							RuntimeVars:          map[string]string{},
						},
					},
				},
			}
			err := k8sClient.Create(ctx, crdSpec)
			Expect(err).Should(BeNil())

			specLookupKey := types.NamespacedName{Name: RestStrategyName, Namespace: RestStrategyNamespace}
			createdSpec := &seederv1alpha1.RestStrategy{}

			// We'll need to retry getting this newly created RestStrategy, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, specLookupKey, createdSpec)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			Expect(createdSpec.Spec.AuthConfig[0].Name).Should(Equal("basic"))
			Expect(len(createdSpec.Spec.Seeders)).Should(Equal(2))
			By("By checking the UpdatedRestStrategy")
			Consistently(func() (string, error) {
				updatedSpec := &seederv1alpha1.RestStrategy{}
				err := k8sClient.Get(ctx, specLookupKey, updatedSpec)
				if err != nil {
					return "", err
				}
				return updatedSpec.Status.Message, nil
			}, duration, interval).Should(ContainSubstring(fmt.Sprintf("failed synced resource: %s in namespace: %s", RestStrategyName, RestStrategyNamespace)))
		})
	})

})
