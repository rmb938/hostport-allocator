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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	intmetav1 "github.com/rmb938/hostport-allocator/apis/meta/v1"
)

type HostPortPhase string

const (
	HostPortPhasePending   HostPortPhase = "Pending"
	HostPortPhaseAllocated HostPortPhase = "Allocated"

	HostPortPhaseDeleting HostPortPhase = "Deleting"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HostPortSpec defines the desired state of HostPort
type HostPortSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The referencing claim
	// +kubebuilder:validation:Optional
	ClaimRef *v1.ObjectReference `json:"claimRef,omitempty"`

	// +kubebuilder:validation:Required
	HostPortClassName string `json:"hostPortClassName"`
}

// HostPortStatus defines the observed state of HostPort
type HostPortStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Resource status conditions
	// +kubebuilder:validation:Optional
	Conditions []intmetav1.Condition `json:"conditions,omitempty"`

	// The port that was allocated by the HostPortClass
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port"`

	// +kubebuilder:validation:Optional
	Phase HostPortPhase `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=hp
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="CLASS",type=string,JSONPath=`.spec.hostPortClassName`,priority=0
// +kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.phase`,priority=0
// +kubebuilder:printcolumn:name="PORT",type=integer,JSONPath=`.status.port`,priority=0

// HostPort is the Schema for the hostports API
type HostPort struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec HostPortSpec `json:"spec,omitempty"`

	// +kubebuilder:validation:Optional
	Status HostPortStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HostPortList contains a list of HostPort
type HostPortList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HostPort `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HostPort{}, &HostPortList{})
}
