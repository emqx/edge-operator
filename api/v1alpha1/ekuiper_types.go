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

// EKuiperSpec defines the desired state of EKuiper
type EKuiperSpec struct {
	EdgePodSpec `json:",inline"`

	//+kubebuilder:default:=1
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=1
	Replicas            *int32                                `json:"replicas,omitempty"`
	EKuiper             corev1.Container                      `json:"ekuiper,omitempty"`
	VolumeClaimTemplate *corev1.PersistentVolumeClaimTemplate `json:"volumeClaimTemplate,omitempty"`
	ServiceTemplate     *corev1.Service                       `json:"serviceTemplate,omitempty"`
}

func (ek *EKuiper) GetComponentType() ComponentType {
	return ComponentTypeEKuiper
}

func (ek *EKuiper) GetEdgePodSpec() EdgePodSpec {
	return ek.Spec.EdgePodSpec
}

func (ek *EKuiper) GetNeuron() *corev1.Container {
	return nil
}

func (ek *EKuiper) GetEKuiper() *corev1.Container {
	return &ek.Spec.EKuiper
}

func (ek *EKuiper) GetVolumeClaimTemplate() *corev1.PersistentVolumeClaimTemplate {
	return ek.Spec.VolumeClaimTemplate
}
func (ek *EKuiper) SetVolumeClaimTemplate(pvc *corev1.PersistentVolumeClaimTemplate) {
	ek.Spec.VolumeClaimTemplate = pvc
}

func (ek *EKuiper) GetServiceTemplate() *corev1.Service {
	return ek.Spec.ServiceTemplate
}
func (ek *EKuiper) SetServiceTemplate(svc *corev1.Service) {
	ek.Spec.ServiceTemplate = svc
}

func (ek *EKuiper) GetReplicas() *int32 {
	return ek.Spec.Replicas
}

func (ek *EKuiper) SetReplicas(replicas int32) {
	ek.Spec.Replicas = &replicas
}

// EKuiperStatus defines the observed state of EKuiper
type EKuiperStatus struct {
	EdgeStatus `json:",inline"`
}

func (n *EKuiper) GetStatus() EdgeStatus {
	return n.Status.EdgeStatus
}

func (n *EKuiper) SetStatus(status *EdgeStatus) {
	n.Status.EdgeStatus = *status
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas

// EKuiper is the Schema for the ekuipers API
type EKuiper struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EKuiperSpec   `json:"spec,omitempty"`
	Status EKuiperStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EKuiperList contains a list of EKuiper
type EKuiperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EKuiper `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EKuiper{}, &EKuiperList{})
}
