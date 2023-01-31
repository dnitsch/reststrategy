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
	Spec              StrategySpec   `json:"spec,omitempty"`
	Status            StrategyStatus `json:"status"`
}
// https://control-plane:8080/apis/dnitsch.net/v1alpha/reststrategies/
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

// StrategyStatus is the status for a RestStrategy resource
type StrategyStatus struct {
	Message string `json:"message"`
	// ...
	// add more fields as and when necessary
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// RestStrategyList is a list of RestStrategy resources
type RestStrategyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RestStrategy `json:"items"`
}
