package k8s_test

import (
	"os"
	"testing"

	"github.com/dnitsch/reststrategy/apis/reststrategy/generated/clientset/versioned"
	"github.com/dnitsch/reststrategy/apis/reststrategy/generated/clientset/versioned/fake"
	"github.com/dnitsch/reststrategy/controller/internal/k8s"
)

type args struct {
	onboardClient versioned.Interface
	namespace     string
}

const (
	TestPhrase         string = "got: %v want: %v"
	TestPhraseWContext string = "failed %s => got: %v want: %v"
)

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

			got, _ := k8s.InitialiseSharedInformerFactory(tt.args.onboardClient, tt.args.namespace, 60)
			if got.Reststrategy().V1alpha1() == nil {
				t.Errorf("InitialiseSharedInformerFactory() got = %v, want not <nil>", got)
			}
		})
	}
}

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
			ns := k8s.GetNamespace(tt.ns)
			if ns != tt.want {
				t.Errorf(TestPhrase, tt.want, ns)
			}
		})
	}
}
