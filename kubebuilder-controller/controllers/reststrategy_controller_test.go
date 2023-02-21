package controllers

import (
	"context"
	"time"

	seederv1alpha1 "github.com/dnitsch/reststrategy/kubebuilder-controller/api/v1alpha1"
	"github.com/dnitsch/reststrategy/seeder"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("RestStrategy controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		RestStrategyName      = "test-rest-strategy-mocked"
		RestStrategyNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

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
						{
							Name:         "customTest",
							AuthStrategy: seeder.CustomToToken,
							Username:     "_",
							Password:     "_r",
							CustomToken: &seeder.CustomToken{
								AuthUrl: "http://127.0.0.1:8090/api/admins/auth-via-email",
								CustomAuthMap: seeder.KvMapVarsAny{
									"email":    "test@example.com",
									"password": "P4s$w0rd123!",
								},
								ResponseKey: "$.token",
								HeaderKey:   "Authorization",
								TokenPrefix: "Admin",
							},
						},
					},
					Seeders: []seeder.Action{
						{
							Name:               "test123",
							Endpoint:           "http://127.0.0.1:8090/api/admins",
							Strategy:           "FIND/PATCH/POST",
							GetEndpointSuffix:  seeder.String("?page=1&perPage=100&sort=-created&filter="),
							FindByJsonPathExpr: "$.items[?(@.email=='test2@example.com')].id",
							AuthMapRef:         "customTest",
							PayloadTemplate: `{"email": "test2@example.com",
								"password": "${password}","passwordConfirm": "${password}",
								"avatar": 7
							}`,
							PatchPayloadTemplate: `{
								"password": "${password}",
								"passwordConfirm": "${password}"
							  }`,
							Variables: seeder.KvMapVarsAny{
								"password": "NewPassdd123",
							},
							RuntimeVars: map[string]string{
								"admin1AvatarId": "$.avatar",
							},
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
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdSpec.Spec.Seeders[0].Name).Should(Equal("test123"))
			By("By checking the RestStrategy")
			Consistently(func() (string, error) {
				err := k8sClient.Get(ctx, specLookupKey, createdSpec)
				if err != nil {
					return "", err
				}
				return createdSpec.Status.Message, nil
			}, duration, interval).Should(Equal(""))
		})
	})

})
