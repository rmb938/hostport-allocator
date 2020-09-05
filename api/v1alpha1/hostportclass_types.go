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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type HostPortClassSpecPool struct {
	// The start port for the pool
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Start int `json:"start"`

	// The end port for the pool
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	End int `json:"end"`
}

// HostPortClassSpec defines the desired state of HostPortClass
type HostPortClassSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	Pools []HostPortClassSpecPool `json:"pools"`
}

// HostPortClassStatus defines the observed state of HostPortClass
type HostPortClassStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=hpcl
// +kubebuilder:storageversion
// +kubebuilder:subresource:status

// HostPortClass is the Schema for the hostportclasses API
type HostPortClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec HostPortClassSpec `json:"spec,omitempty"`

	// +kubebuilder:validation:Optional
	Status HostPortClassStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HostPortClassList contains a list of HostPortClass
type HostPortClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HostPortClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HostPortClass{}, &HostPortClassList{})
}
