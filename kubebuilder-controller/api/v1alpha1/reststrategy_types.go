/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/dnitsch/reststrategy/seeder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:object:generate=true

type RestStrategy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StrategySpec   `json:"spec,omitempty"`
	Status            StrategyStatus `json:"status,omitempty"`
}

type StrategySpec struct {
	AuthConfig []seeder.AuthConfig `json:"auth"`
	Seeders    []seeder.Action     `json:"seed"`
}

// StrategyStatus is the status for a RestStrategy resource
type StrategyStatus struct {
	Message string `json:"message"`
	// ...
	// add more fields as and when necessary
}

//+kubebuilder:object:root=true
// +kubebuilder:object:generate=true

// RestStrategyList is a list of RestStrategy resources
type RestStrategyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RestStrategy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RestStrategy{}, &RestStrategyList{})
}
