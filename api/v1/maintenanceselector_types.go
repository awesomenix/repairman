/*

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MaintenanceSelectorSpec defines the desired state of MaintenanceSelector
type MaintenanceSelectorSpec struct {
}

// MaintenanceSelectorStatus defines the observed state of MaintenanceSelector
type MaintenanceSelectorStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MaintenanceSelector is the Schema for the maintenanceselectors API
type MaintenanceSelector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MaintenanceSelectorSpec   `json:"spec,omitempty"`
	Status MaintenanceSelectorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MaintenanceSelectorList contains a list of MaintenanceSelector
type MaintenanceSelectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MaintenanceSelector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MaintenanceSelector{}, &MaintenanceSelectorList{})
}
