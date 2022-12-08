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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NeuronSpec defines the desired state of Neuron
type NeuronSpec struct {
	EdgePodSpec `json:",inline"`

	Neuron *corev1.Container `json:"neuron,omitempty"`

	ServiceTemplate     *corev1.Service               `json:"serviceTemplate,omitempty"`
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

func (n *Neuron) GetEdgePodSpec() EdgePodSpec {
	return n.Spec.EdgePodSpec
}
func (n *Neuron) SetEdgePodSpec(spec EdgePodSpec) {
	n.Spec.EdgePodSpec = spec
}

func (n *Neuron) GetNeuron() *corev1.Container {
	return n.Spec.Neuron
}
func (n *Neuron) SetNeuron(container *corev1.Container) {
	n.Spec.Neuron = container
}

func (n *Neuron) GetEKuiper() *corev1.Container {
	return nil
}
func (n *Neuron) SetEKuiper(container *corev1.Container) {
}

func (n *Neuron) GetVolumeClaimTemplate() *corev1.PersistentVolumeClaim {
	return n.Spec.VolumeClaimTemplate
}
func (n *Neuron) SetVolumeClaimTemplate(claim *corev1.PersistentVolumeClaim) {
	n.Spec.VolumeClaimTemplate = claim
}

func (n *Neuron) GetServiceTemplate() *corev1.Service {
	return n.Spec.ServiceTemplate
}
func (n *Neuron) SetServiceTemplate(service *corev1.Service) {
	n.Spec.ServiceTemplate = service
}

// NeuronStatus defines the observed state of Neuron
type NeuronStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Neuron is the Schema for the neurons API
type Neuron struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NeuronSpec   `json:"spec,omitempty"`
	Status NeuronStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NeuronList contains a list of Neuron
type NeuronList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Neuron `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Neuron{}, &NeuronList{})
}
