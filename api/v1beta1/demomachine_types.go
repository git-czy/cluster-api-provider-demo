/*
Copyright 2022.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// ClusterFinalizer allows ReconcileDemoMachine to clean up resources associated with metalNode before
	// removing it from the apiserver.
	MachineFinalizer = "demomachine.infrastructure.cluster.x-k8s.io"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DemoMachineSpec defines the desired state of DemoMachine
type DemoMachineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ProviderID will be the metal node recourse uid,it's unique
	// +optional
	ProviderID string `json:"providerID,omitempty"`
}

// DemoMachineStatus defines the observed state of DemoMachine
type DemoMachineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Ready denotes that the machine (bare metal) is ready
	// +optional
	Ready bool `json:"ready"`

	//Bootstrapped means that the machine already has bootstrapped
	// +optional
	Bootstrapped bool `json:"bootstrapped"`

	// Addresses contains the associated addresses for the demo machine.
	// +optional
	Addresses []clusterv1.MachineAddress `json:"addresses,omitempty"`

	// Conditions defines current service state of the DemoMachine.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DemoMachine is the Schema for the demomachines API
type DemoMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DemoMachineSpec   `json:"spec,omitempty"`
	Status DemoMachineStatus `json:"status,omitempty"`
}

// GetConditions returns the set of conditions for this object.
func (m *DemoMachine) GetConditions() clusterv1.Conditions {
	return m.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (m *DemoMachine) SetConditions(conditions clusterv1.Conditions) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// DemoMachineList contains a list of DemoMachine
type DemoMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DemoMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DemoMachine{}, &DemoMachineList{})
}
