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

package controllers

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/emqx/edge-operator/internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("NeuronEX controller", func() {
	var ins, new *edgev1alpha1.NeuronEX

	BeforeEach(func() {
		ins = &edgev1alpha1.NeuronEX{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "neuronex",
				Namespace: "default",
				Annotations: map[string]string{
					"foo": "bar",
				},
			},
			Spec: edgev1alpha1.NeuronEXSpec{
				EKuiper: corev1.Container{
					Name:  "ekuiper",
					Image: "lfedge/ekuiper:1.8-slim-python",
				},
				Neuron: corev1.Container{
					Name:  "neuron",
					Image: "emqx/neuron:2.3",
				},
			},
		}
		ins.Default()
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(Succeed())
	})
	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, ins.DeepCopy())).Should(Succeed())
	})

	It("should create configMap", func() {
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      internal.GetResNameOnPanic(ins, internal.EKuiperToolConfig),
				Namespace: ins.Namespace,
			},
		}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKeyFromObject(configMap), configMap)
		}, timeout, interval).Should(Succeed())

		// metadata
		Expect(configMap.ObjectMeta.Labels).Should(Equal(ins.Labels))
		Expect(configMap.ObjectMeta.Annotations).Should(HaveKeyWithValue("foo", "bar"))
		// data
		Expect(configMap.Data).Should(HaveKey("neuronStream.json"))
	})

	It("should create deployment", func() {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ins.GetResName(),
				Namespace: ins.GetNamespace(),
			},
		}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
		}, timeout, interval).Should(Succeed())

		// metadata
		Expect(deployment.ObjectMeta.Labels).Should(Equal(ins.Labels))
		Expect(deployment.ObjectMeta.Annotations).Should(HaveKeyWithValue("foo", "bar"))

		// neuron EX have three containers
		Expect(deployment.Spec.Template.Spec.Containers).Should(HaveLen(3))
		// neuron container
		Expect(deployment.Spec.Template.Spec.Containers[0].Env).Should(ConsistOf([]corev1.EnvVar{
			{Name: "LOG_CONSOLE", Value: "true"},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[0].Name).Should(Equal(ins.Spec.Neuron.Name))
		Expect(deployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(ins.Spec.Neuron.Image))
		Expect(deployment.Spec.Template.Spec.Containers[0].Ports).Should(ConsistOf([]corev1.ContainerPort{
			{Name: "neuron", ContainerPort: 7000, Protocol: corev1.ProtocolTCP},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts).Should(ConsistOf([]corev1.VolumeMount{
			{Name: "shared-tmp", MountPath: "/tmp"},
			{Name: "neuron-data", MountPath: "/opt/neuron/persistence"},
		}))
		// ekuiper container
		Expect(deployment.Spec.Template.Spec.Containers[1].Name).Should(Equal(ins.Spec.EKuiper.Name))
		Expect(deployment.Spec.Template.Spec.Containers[1].Image).Should(Equal(ins.Spec.EKuiper.Image))
		Expect(deployment.Spec.Template.Spec.Containers[1].Env).Should(ConsistOf([]corev1.EnvVar{
			{Name: "KUIPER__BASIC__RESTPORT", Value: "9081"},
			{Name: "KUIPER__BASIC__IGNORECASE", Value: "false"},
			{Name: "KUIPER__BASIC__CONSOLELOG", Value: "true"},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[1].Ports).Should(ConsistOf([]corev1.ContainerPort{
			{Name: "ekuiper", ContainerPort: 9081, Protocol: corev1.ProtocolTCP},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts).Should(ConsistOf([]corev1.VolumeMount{
			{Name: "shared-tmp", MountPath: "/tmp"},
			{Name: "ekuiper-data", MountPath: "/kuiper/data"},
			{Name: "ekuiper-plugins", MountPath: "/kuiper/plugins/portable"},
		}))
		// ekuiper tool container
		Expect(deployment.Spec.Template.Spec.Containers[2].Name).Should(Equal("ekuiper-tool"))
		Expect(deployment.Spec.Template.Spec.Containers[2].Image).Should(Equal("lfedge/ekuiper-kubernetes-tool:1.8"))
		Expect(deployment.Spec.Template.Spec.Containers[2].VolumeMounts).Should(ConsistOf([]corev1.VolumeMount{
			{Name: "ekuiper-tool-config", MountPath: "/kuiper-kubernetes-tool/sample", ReadOnly: true},
		}))

		// volume
		Expect(deployment.Spec.Template.Spec.Volumes).Should(ConsistOf([]corev1.Volume{
			{Name: "shared-tmp", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			{Name: "neuron-data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			{Name: "ekuiper-data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			{Name: "ekuiper-plugins", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			{Name: "ekuiper-tool-config", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: internal.GetResNameOnPanic(ins, internal.EKuiperToolConfig)},
				DefaultMode:          &[]int32{corev1.ConfigMapVolumeSourceDefaultMode}[0],
			}}},
		}))
	})

	Describe("Update NeuronEX", func() {
		JustBeforeEach(func() {
			new = ins.DeepCopy()
			new.Spec.VolumeClaimTemplate = &corev1.PersistentVolumeClaim{
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("8Mi")},
					},
				},
			}
			new.Spec.ServiceTemplate = &corev1.Service{}
			new.Default()
			Expect(k8sClient.Patch(ctx, new, client.MergeFrom(ins))).Should(Succeed())
		})

		It("should create service", func() {
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      new.Spec.ServiceTemplate.Name,
					Namespace: new.GetNamespace(),
				},
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(service), service)
			}, timeout, interval).Should(Succeed())

			Expect(service.Spec.Ports).Should(ConsistOf([]corev1.ServicePort{
				{Name: "neuron", Port: 7000, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(7000)},
				{Name: "ekuiper", Port: 9081, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(9081)},
			}))
		})

		It("should create three PVC", func() {
			pvcList := &corev1.PersistentVolumeClaimList{}
			Expect(k8sClient.List(ctx, pvcList, client.InNamespace(new.GetNamespace()), client.MatchingLabels(new.GetLabels()))).Should(Succeed())

			Expect(pvcList.Items).Should(HaveLen(3))
			for _, pvc := range pvcList.Items {
				Expect(pvc.Name).Should(BeElementOf([]string{
					new.GetResName() + "-neuron-data",
					new.GetResName() + "-ekuiper-data",
					new.GetResName() + "-ekuiper-plugins",
				}))
				Expect(pvc.Labels).Should(Equal(ins.GetLabels()))
				Expect(pvc.Annotations).Should(HaveKeyWithValue("foo", "bar"))

				Expect(pvc.Spec.AccessModes).Should(Equal(new.Spec.VolumeClaimTemplate.Spec.AccessModes))
				Expect(pvc.Spec.Resources).Should(Equal(new.Spec.VolumeClaimTemplate.Spec.Resources))
			}
		})

		It("should update deployment", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      new.GetResName(),
					Namespace: new.GetNamespace(),
				},
			}
			// Make sure deployment already updated
			Eventually(func() []corev1.Volume {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
				if err != nil {
					return nil
				}
				return deployment.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(ConsistOf([]corev1.Volume{
				{Name: "shared-tmp", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "neuron-data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: new.Spec.VolumeClaimTemplate.Name + "-neuron-data"}}},
				{Name: "ekuiper-data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: new.Spec.VolumeClaimTemplate.Name + "-ekuiper-data"}}},
				{Name: "ekuiper-plugins", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: new.Spec.VolumeClaimTemplate.Name + "-ekuiper-plugins"}}},
				{Name: "ekuiper-tool-config", VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: internal.GetResNameOnPanic(new, internal.EKuiperToolConfig)},
						DefaultMode:          &[]int32{corev1.ConfigMapVolumeSourceDefaultMode}[0],
					},
				}},
			}))
		})
	})
})
