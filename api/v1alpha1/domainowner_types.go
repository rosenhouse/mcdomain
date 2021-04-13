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

// DomainOwnerSpec defines the desired state of DomainOwner
type DomainOwnerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Domain is a DNS suffix
	Domain string `json:"domain,omitempty"`

	// Owners specify which Clusters may serve the Domain
	Owners *metav1.LabelSelector `json:"owners,omitempty"`
}

// DomainOwnerStatus defines the observed state of DomainOwner
type DomainOwnerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// DomainOwner is the Schema for the domainowners API
type DomainOwner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DomainOwnerSpec   `json:"spec,omitempty"`
	Status DomainOwnerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DomainOwnerList contains a list of DomainOwner
type DomainOwnerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DomainOwner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DomainOwner{}, &DomainOwnerList{})
}
