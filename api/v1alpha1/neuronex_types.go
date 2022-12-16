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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NeuronEXSpec defines the desired state of NeuronEX
type NeuronEXSpec struct {
	EdgePodSpec `json:",inline"`

	Neuron  corev1.Container `json:"neuron,omitempty"`
	EKuiper corev1.Container `json:"ekuiper,omitempty"`

	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
	ServiceTemplate     *corev1.Service               `json:"serviceTemplate,omitempty"`
}

func (n *NeuronEX) GetComponentType() ComponentType {
	return ComponentTypeNeuronEx
}

func (n *NeuronEX) GetResName() string {
	return fmt.Sprintf("%s-%s", n.GetName(), ComponentTypeNeuronEx)
}

func (n *NeuronEX) GetEdgePodSpec() EdgePodSpec {
	return n.Spec.EdgePodSpec
}

func (n *NeuronEX) GetNeuron() *corev1.Container {
	return &n.Spec.Neuron
}

func (n *NeuronEX) GetEKuiper() *corev1.Container {
	return &n.Spec.EKuiper
}

func (n *NeuronEX) GetVolumeClaimTemplate() *corev1.PersistentVolumeClaim {
	return n.Spec.VolumeClaimTemplate
}

func (n *NeuronEX) GetServiceTemplate() *corev1.Service {
	return n.Spec.ServiceTemplate
}

// NeuronEXStatus defines the observed state of NeuronEX
type NeuronEXStatus struct {
	EdgeStatus `json:",inline"`
}

func (n *NeuronEX) GetStatus() EdgeStatus {
	return n.Status.EdgeStatus
}

func (n *NeuronEX) SetStatus(status EdgeStatus) {
	n.Status.EdgeStatus = status
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=neuronexs,shortName=nex

// NeuronEX is the Schema for the neuronexs API
type NeuronEX struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NeuronEXSpec   `json:"spec,omitempty"`
	Status NeuronEXStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NeuronEXList contains a list of NeuronEX
type NeuronEXList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NeuronEX `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NeuronEX{}, &NeuronEXList{})
}
