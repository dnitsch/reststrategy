package k8sutils

import (
	"os"
	"testing"

	"github.com/dnitsch/reststrategy/apis/generated/clientset/versioned"
	"github.com/dnitsch/reststrategy/apis/generated/clientset/versioned/fake"
	"github.com/dnitsch/reststrategy/controller/internal/testutils"
)

type args struct {
	onboardClient versioned.Interface
	namespace     string
}

func TestInitialiseSharedInformerFactory_withNamespace(t *testing.T) {

	tests := []struct {
		name    string
		args    args
		envFunc func()
		want    string
	}{
		{
			name: "Initialise with namespace argument",
			args: args{
				onboardClient: &fake.Clientset{},
				namespace:     "METHOD_ARGUMENT",
			},
			envFunc: func() {
				os.Clearenv()
			},
			want: "METHOD_ARGUMENT",
		},
		{
			name: "Initialise with environment variable",
			args: args{
				onboardClient: &fake.Clientset{},
				namespace:     "",
			},
			envFunc: func() {
				os.Clearenv()
				os.Setenv("POD_NAMESPACE", "ENVIRONMENT_VARIABLE")
			},
			want: "ENVIRONMENT_VARIABLE",
		},
		{
			name: "namespace argument is chosen over environment variable",
			args: args{
				onboardClient: &fake.Clientset{},
				namespace:     "METHOD_ARGUMENT",
			},
			envFunc: func() {
				os.Clearenv()
				os.Setenv("POD_NAMESPACE", "ENVIRONMENT_VARIABLE")
			},
			want: "METHOD_ARGUMENT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.envFunc()

			got, _ := initialiseSharedInformerFactory(tt.args.onboardClient, tt.args.namespace, 60)
			if got.Reststrategy().V1alpha1() == nil {
				t.Errorf("InitialiseSharedInformerFactory() got = %v, want not <nil>", got)
			}
		})
	}
}

// func TestInitialiseSharedInformerFactory_withoutNamespace(t *testing.T) {

// 	tests := []struct {
// 		name string
// 		args args
// 		want error
// 	}{
// 		{
// 			name: "Initialise with empty namespace argument and no environment variable set",
// 			args: args{
// 				onboardClient: &fake.Clientset{},
// 				namespace:     "",
// 			},
// 			want: errors.New("either --namespace arg must be provided or POD_NAMESPACE env variable must be present"),
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			_, err := InitialiseSharedInformerFactory(tt.args.onboardClient, tt.args.namespace, 60)
// 			if err == nil {
// 				t.Errorf("Expected InitialiseSharedInformerFactory() to return Error")
// 			}
// 			if err.Error() != tt.want.Error() {
// 				t.Errorf("expected error to be: %v", tt.want.Error())
// 			}
// 		})
// 	}
// }

func Test_GetNs(t *testing.T) {
	tests := []struct {
		name    string
		ns      string
		envFunc func()
		want    string
	}{
		{
			name:    "param not empty",
			ns:      "test",
			envFunc: func() { os.Clearenv() },
			want:    "test",
		},
		{
			name: "param passed via env var",
			ns:   "",
			envFunc: func() {
				os.Clearenv()
				os.Setenv("POD_NAMESPACE", "env-var-ns")
			},
			want: "env-var-ns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.envFunc()
			ns := getNamespace(tt.ns)
			if ns != tt.want {
				t.Errorf(testutils.TestPhrase, tt.want, ns)
			}
		})
	}
}
