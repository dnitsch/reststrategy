package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RestStrategy is a specification for a RestStrategy resource
type RestStrategy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RestStrategySpec `json:"spec,omitempty"`
	// Status            RestStrategyStatus `json:"status"`
}

// RestStrategySpec is the spec for a RestStrategy resource
type RestStrategySpec struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RestStrategyList is a list of RestStrategy resources
type RestStrategyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RestStrategy `json:"items"`
}
