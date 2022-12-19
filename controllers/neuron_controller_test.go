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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Neuron controller", func() {
	var ins, new *edgev1alpha1.Neuron

	BeforeEach(func() {
		ins = &edgev1alpha1.Neuron{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "neuron",
				Namespace: "default",
				Annotations: map[string]string{
					"foo": "bar",
				},
			},
			Spec: edgev1alpha1.NeuronSpec{
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

	It("check custom resources status", func() {
		// envTest does not create pods via deployment, they need to be created manually
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "fake-",
				Namespace:    ins.GetNamespace(),
				Labels:       ins.GetLabels(),
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "busybox", Image: "busybox"},
				},
			},
		}
		Expect(k8sClient.Create(ctx, pod)).Should(Succeed())
		Eventually(func() corev1.PodPhase {
			got := ins.DeepCopy()
			_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(got), got)
			return got.Status.Phase
		}, timeout, interval).Should(BeZero())

		pod.Status.Phase = corev1.PodRunning
		Expect(k8sClient.Status().Update(ctx, pod)).Should(Succeed())
		Eventually(func() corev1.PodPhase {
			got := ins.DeepCopy()
			_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(got), got)
			return got.Status.Phase
		}, timeout, interval).Should(Equal(corev1.PodRunning))
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
		Expect(deployment.Spec.Template.Spec.Containers).Should(HaveLen(1))
		// neuron container
		Expect(deployment.Spec.Template.Spec.Containers[0].Env).Should(ConsistOf([]corev1.EnvVar{
			{Name: "LOG_CONSOLE", Value: "1"},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[0].Name).Should(Equal(ins.Spec.Neuron.Name))
		Expect(deployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(ins.Spec.Neuron.Image))
		Expect(deployment.Spec.Template.Spec.Containers[0].Ports).Should(ConsistOf([]corev1.ContainerPort{
			{Name: "neuron", ContainerPort: 7000, Protocol: corev1.ProtocolTCP},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts).Should(ConsistOf([]corev1.VolumeMount{
			{Name: "neuron-data", MountPath: "/opt/neuron/persistence"},
		}))

		// volume
		Expect(deployment.Spec.Template.Spec.Volumes).Should(ConsistOf([]corev1.Volume{
			{Name: "neuron-data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
		}))
	})

	Describe("Update Neuron", func() {
		JustBeforeEach(func() {
			new = ins.DeepCopy()
			new.Annotations["updated"] = "done"
			new.Spec.VolumeClaimTemplate = &corev1.PersistentVolumeClaim{
				Spec: corev1.PersistentVolumeClaimSpec{
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
			}))
		})

		It("should create three PVC", func() {
			pvcList := &corev1.PersistentVolumeClaimList{}
			Expect(k8sClient.List(ctx, pvcList, client.InNamespace(new.GetNamespace()), client.MatchingLabels(new.GetLabels()))).Should(Succeed())

			Expect(pvcList.Items).Should(HaveLen(1))
			for _, pvc := range pvcList.Items {
				Expect(pvc.Name).Should(BeElementOf([]string{
					new.GetResName() + "-neuron-data",
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
				{Name: "neuron-data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: new.Spec.VolumeClaimTemplate.Name + "-neuron-data"}}},
			}))
		})
	})
})
