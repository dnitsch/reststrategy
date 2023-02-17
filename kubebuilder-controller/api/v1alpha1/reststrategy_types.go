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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

//+kubebuilder:object:root=true
//+kubebuilder:object:generate=true

type RestStrategy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StrategySpec   `json:"spec,omitempty"`
	Status            StrategyStatus `json:"status,omitempty"`
}

// type AuthConfig struct {
// 	Name              string
// 	seeder.AuthConfig `json:",inline"`
// }

// type SeederConfig struct {
// 	Name          string `json:"name"`
// 	seeder.Action `json:",inline"`
// }

type StrategySpec struct {
	AuthConfig []seeder.AuthConfig `json:"auth"`
	Seeders    []seeder.Action     `json:"seed"`
}

// +kubebuilder:subresource:status
//
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
