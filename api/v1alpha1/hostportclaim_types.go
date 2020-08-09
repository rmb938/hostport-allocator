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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	intmetav1 "github.com/rmb938/hostport-allocator/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HostPortClaimSpec defines the desired state of HostPortClaim
type HostPortClaimSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The host port class
	// +kubebuilder:validation:Required
	HostPortClassName string `json:"hostPortClassName"`

	// The binding reference to the HostPort backing this claim
	// +kubebuilder:validation:Optional
	HostPortName string `json:"hostPortName"`
}

type HostPortClaimStatusPhase string

// HostPortClaimStatus defines the observed state of HostPortClaim
type HostPortClaimStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Resource status conditions
	// +kubebuilder:validation:Optional
	Conditions []intmetav1.Condition `json:"conditions,omitempty"`

	// +kubebuilder:validation:Optional
	Phase HostPortClaimStatusPhase `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=hpc
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.phase`,priority=0
// +kubebuilder:printcolumn:name="HOSTPORTCLASS",type=string,JSONPath=`.spec.hostPortClassName`,priority=0

// HostPortClaim is the Schema for the hostportclaims API
type HostPortClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec HostPortClaimSpec `json:"spec,omitempty"`

	// +kubebuilder:validation:Optional
	Status HostPortClaimStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HostPortClaimList contains a list of HostPortClaim
type HostPortClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HostPortClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HostPortClaim{}, &HostPortClaimList{})
}
