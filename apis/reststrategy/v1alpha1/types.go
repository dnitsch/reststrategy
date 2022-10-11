package v1alpha1

import (
	"github.com/dnitsch/reststrategy/seeder/pkg/rest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// RestStrategy is a specification for a RestStrategy resource
type RestStrategy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StrategySpec `json:"spec,omitempty"`
}

type AuthConfig struct {
	Name string `json:"name"`
	rest.AuthConfig
}

type SeederConfig struct {
	Name string `json:"name"`
	rest.Action
}

type StrategySpec struct {
	AuthConfig []AuthConfig   `json:"auth"`
	Seeders    []SeederConfig `json:"seed"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// RestStrategyList is a list of RestStrategy resources
type RestStrategyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RestStrategy `json:"items"`
}
