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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Neuron validate webhook", func() {
	var ins *Neuron

	BeforeEach(func() {
		ins = new(Neuron)
		ins = &Neuron{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "neuron",
				Namespace: "default",
			},
		}
	})

	AfterEach(func() {
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ins), ins); err == nil {
			Expect(k8sClient.Delete(ctx, ins.DeepCopy())).Should(Succeed())
		}
	})

	It("should block if neuron name is empty", func() {
		ins.Spec.Neuron.Name = ""
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(MatchError(matchError(ins, "neuron container name is empty")))
	})

	It("should block if neuron image is empty", func() {
		ins.Spec.Neuron.Name = "neuron"
		ins.Spec.Neuron.Image = ""
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(MatchError(matchError(ins, "neuron container image is empty")))
	})

	It("should block update if neuron name is empty", func() {
		ins.Spec.Neuron.Name = "neuron"
		ins.Spec.Neuron.Image = "emqx/neuron:2.3"
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(Succeed())

		new := ins.DeepCopy()
		new.Spec.Neuron.Name = ""
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(MatchError(matchError(ins, "neuron container name is empty")))

		new.Spec.Neuron.Name = "neuron"
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(Succeed())
	})

	It("should block update if neuron image is empty", func() {
		ins.Spec.Neuron.Name = "neuron"
		ins.Spec.Neuron.Image = "emqx/neuron:2.3"
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(Succeed())

		new := ins.DeepCopy()
		new.Spec.Neuron.Image = ""
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(MatchError(matchError(ins, "neuron container image is empty")))

		new.Spec.Neuron.Image = "emqx/neuron:latest"
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(Succeed())
	})
})

var _ = Describe("Neuron default webhook", func() {
	var ins *Neuron

	BeforeEach(func() {
		ins = &Neuron{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "neuron",
				Namespace: "default",
				Labels: map[string]string{
					"foo": "bar",
				},
				Annotations: map[string]string{
					"foo": "bar",
				},
			},
			Spec: NeuronSpec{
				Neuron: corev1.Container{
					Name:  "neuron",
					Image: "emqx/neuron:2.3",
					Env: []corev1.EnvVar{
						{Name: "foo", Value: "bar"},
					},
					Ports: []corev1.ContainerPort{
						{
							Name:          "fake-neuron",
							ContainerPort: 1234,
							Protocol:      corev1.ProtocolTCP,
						},
					},
				},
				VolumeClaimTemplate: &corev1.PersistentVolumeClaim{},
				ServiceTemplate:     &corev1.Service{},
			},
		}
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, ins.DeepCopy())).Should(Succeed())
	})

	It("check default values", func() {
		got := &Neuron{}
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(ins), got)).Should(Succeed())

		Expect(got.Labels).Should(HaveKeyWithValue(ManagerByKey, "edge-operator"))
		Expect(got.Labels).Should(HaveKeyWithValue(InstanceKey, got.GetName()))
		Expect(got.Labels).Should(HaveKeyWithValue(ComponentKey, string(got.GetComponentType())))
		Expect(got.Labels).Should(HaveKeyWithValue("foo", "bar"))

		Expect(got.Annotations).Should(HaveKeyWithValue("foo", "bar"))

		Expect(got.GetNeuron().Env).Should(ConsistOf([]corev1.EnvVar{
			{Name: "foo", Value: "bar"},
			{Name: "LOG_CONSOLE", Value: "true"},
		}))
		Expect(got.GetNeuron().Ports).Should(ConsistOf([]corev1.ContainerPort{
			{Name: "fake-neuron", ContainerPort: 1234, Protocol: corev1.ProtocolTCP},
			{Name: "neuron", ContainerPort: 7000, Protocol: corev1.ProtocolTCP},
		}))
		Expect(got.GetNeuron().ReadinessProbe).Should(Equal(&corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			FailureThreshold:    12,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "",
					Port: intstr.FromInt(7000),
				},
			},
		}))
		Expect(got.GetNeuron().LivenessProbe).Should(Equal(&corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			FailureThreshold:    12,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "",
					Port: intstr.FromInt(7000),
				},
			},
		}))

		Expect(got.GetServiceTemplate().Name).Should(Equal(got.GetResName()))
		Expect(got.GetServiceTemplate().Namespace).Should(Equal(got.GetNamespace()))
		Expect(got.GetServiceTemplate().Annotations).Should(Equal(got.GetAnnotations()))
		Expect(got.GetServiceTemplate().Labels).Should(Equal(got.Labels))

		Expect(got.GetServiceTemplate().Spec.Selector).Should(Equal(got.Labels))
		Expect(got.GetServiceTemplate().Spec.Ports).Should(ConsistOf([]corev1.ServicePort{
			{Name: "neuron", Port: 7000, Protocol: corev1.ProtocolTCP, TargetPort: intstr.Parse("7000")},
			{Name: "fake-neuron", Port: 1234, Protocol: corev1.ProtocolTCP, TargetPort: intstr.Parse("1234")},
		}))

		Expect(got.GetVolumeClaimTemplate().Name).Should(Equal(got.GetResName()))
		Expect(got.GetVolumeClaimTemplate().Namespace).Should(Equal(got.GetNamespace()))
		Expect(got.GetVolumeClaimTemplate().Annotations).Should(Equal(got.GetAnnotations()))
		Expect(got.GetServiceTemplate().Annotations).Should(Equal(got.GetAnnotations()))
	})
})
