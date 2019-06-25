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

// MaintenanceState defined below
type MaintenanceState string

const (
	// Pending initial state of maintenance update by actor
	Pending MaintenanceState = "Pending"
	// Approved mainteanance request is approved updated by operator
	Approved MaintenanceState = "Approved"
	// InProgress mainteanance request is in progress, updated by actor
	InProgress MaintenanceState = "InProgress"
	// Completed mainteanance request is completed, updated by actor
	Completed MaintenanceState = "Completed"
)

// MaintenanceRequestSpec defines the desired state of MaintenanceRequest
type MaintenanceRequestSpec struct {
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:Enum={Pending,Approved,InProgress,Completed}
	State MaintenanceState `json:"state,omitempty"`
	// +kubebuilder:validation:Enum=Node
	Type string `json:"type,omitempty"`
}

// MaintenanceRequestStatus defines the observed state of MaintenanceRequest
type MaintenanceRequestStatus struct {
}

// MaintenanceRequest is the Schema for the maintenancerequests API
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=maintenancerequests,scope=Cluster
// +kubebuilder:subresource:status
type MaintenanceRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MaintenanceRequestSpec   `json:"spec,omitempty"`
	Status MaintenanceRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MaintenanceRequestList contains a list of MaintenanceRequest
type MaintenanceRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MaintenanceRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MaintenanceRequest{}, &MaintenanceRequestList{})
}
