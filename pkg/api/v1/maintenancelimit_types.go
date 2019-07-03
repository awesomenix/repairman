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

// MaintenanceLimitSpec defines the desired state of MaintenanceLimit
type MaintenanceLimitSpec struct {
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=5
	Limit uint `json:"limit,omitempty"`
}

// MaintenanceLimitStatus defines the observed state of MaintenanceLimit
type MaintenanceLimitStatus struct {
	// Limit dynamically updated by operator based on current cluster state
	Limit uint `json:"limit,omitempty"`
}

// MaintenanceLimit is the Schema for the maintenancelimits API
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=maintenancelimits,scope=Cluster
// +kubebuilder:subresource:status
type MaintenanceLimit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MaintenanceLimitSpec   `json:"spec,omitempty"`
	Status MaintenanceLimitStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MaintenanceLimitList contains a list of MaintenanceLimit
type MaintenanceLimitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MaintenanceLimit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MaintenanceLimit{}, &MaintenanceLimitList{})
}
