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

var _ = Describe("NeuronEX validate webhook", func() {
	var ins *NeuronEX

	BeforeEach(func() {
		ins = new(NeuronEX)
		ins = &NeuronEX{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "neuronex",
				Namespace: "default",
			},
		}
	})

	AfterEach(func() {
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ins), ins); err == nil {
			Expect(k8sClient.Delete(ctx, ins.DeepCopy())).Should(Succeed())
		}
	})

	It("should block if container name is empty", func() {
		ins.Spec.Neuron.Name = ""
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(MatchError(matchError(ins, "neuron container name is empty")))

		ins.Spec.Neuron.Name = "neuron"
		ins.Spec.Neuron.Image = "emqx/neuron:2.3"
		ins.Spec.EKuiper.Name = ""
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(MatchError(matchError(ins, "ekuiper container name is empty")))
	})

	It("should block if container image is empty", func() {
		ins.Spec.Neuron.Name = "neuron"
		ins.Spec.Neuron.Image = ""
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(MatchError(matchError(ins, "neuron container image is empty")))

		ins.Spec.Neuron.Name = "neuron"
		ins.Spec.Neuron.Image = "emqx/neuron:2.3"
		ins.Spec.EKuiper.Name = "ekuiper"
		ins.Spec.EKuiper.Image = ""
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(MatchError(matchError(ins, "ekuiper container image is empty")))
	})

	It("should block update if container name is empty", func() {
		ins.Spec.Neuron.Name = "neuron"
		ins.Spec.Neuron.Image = "emqx/neuron:2.3"
		ins.Spec.EKuiper.Name = "ekuiper"
		ins.Spec.EKuiper.Image = "lfedge/ekuiper:1.8-slim-python"
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(Succeed())

		new := ins.DeepCopy()
		new.Spec.Neuron.Name = ""
		new.Spec.EKuiper.Name = "ekuiper"
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(MatchError(matchError(ins, "neuron container name is empty")))

		new.Spec.Neuron.Name = "neuron"
		new.Spec.EKuiper.Name = ""
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(MatchError(matchError(ins, "ekuiper container name is empty")))

		new.Spec.Neuron.Name = "neuron"
		new.Spec.EKuiper.Name = "ekuiper"
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(Succeed())
	})

	It("should block update if container image is empty", func() {
		ins.Spec.Neuron.Name = "neuron"
		ins.Spec.Neuron.Image = "emqx/neuron:2.3"
		ins.Spec.EKuiper.Name = "ekuiper"
		ins.Spec.EKuiper.Image = "lfedge/ekuiper:1.8-slim-python"
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(Succeed())

		new := ins.DeepCopy()
		new.Spec.Neuron.Image = ""
		new.Spec.EKuiper.Image = "lfedge/ekuiper:1.8-slim-python"
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(MatchError(matchError(ins, "neuron container image is empty")))

		new.Spec.Neuron.Image = "emqx/neuron:2.3"
		new.Spec.EKuiper.Image = ""
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(MatchError(matchError(ins, "ekuiper container image is empty")))

		new.Spec.Neuron.Image = "emqx/neuron:2.3"
		new.Spec.EKuiper.Image = "lfedge/ekuiper:latest-slim-python"
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(Succeed())
	})

	It("should block create if ekuiper image is not slim or slim-python", func() {
		ins.Spec.Neuron.Name = "neuron"
		ins.Spec.Neuron.Image = "emqx/neuron:2.3"
		ins.Spec.EKuiper.Name = "ekuiper"
		ins.Spec.EKuiper.Image = "lfedge/ekuiper:1.8"
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(MatchError(matchError(ins, "ekuiper container image must be slim or slim-python")))
		ins.Spec.EKuiper.Image = "lfedge/ekuiper:1.8-slim"
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(Succeed())
	})

	It("should block update if ekuiper image is slim-python or not slim-python", func() {
		ins.Spec.Neuron.Name = "neuron"
		ins.Spec.Neuron.Image = "emqx/neuron:2.3"
		ins.Spec.EKuiper.Name = "ekuiper"
		ins.Spec.EKuiper.Image = "lfedge/ekuiper:1.8-slim"
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(Succeed())

		new := ins.DeepCopy()
		new.Spec.EKuiper.Image = "lfedge/ekuiper:latest"
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(MatchError(matchError(ins, "ekuiper container image must be slim or slim-python")))
		new.Spec.EKuiper.Image = "lfedge/ekuiper:latest-slim-python"
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(Succeed())
	})
})

var _ = Describe("NeuronEX default webhook", func() {
	var ins *NeuronEX

	BeforeEach(func() {
		ins = &NeuronEX{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "neuronex",
				Namespace: "default",
				Labels: map[string]string{
					"foo": "bar",
				},
				Annotations: map[string]string{
					"foo": "bar",
				},
			},
			Spec: NeuronEXSpec{
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
				EKuiper: corev1.Container{
					Name:  "ekuiper",
					Image: "lfedge/ekuiper:1.8-slim-python",
					Env: []corev1.EnvVar{
						{Name: "foo", Value: "bar"},
					},
					Ports: []corev1.ContainerPort{
						{
							Name:          "fake-ekuiper",
							ContainerPort: 5678,
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
		got := &NeuronEX{}
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

		Expect(got.GetEKuiper().Env).Should(ConsistOf([]corev1.EnvVar{
			{Name: "foo", Value: "bar"},
			{Name: "KUIPER__BASIC__RESTPORT", Value: "9081"},
			{Name: "KUIPER__BASIC__IGNORECASE", Value: "false"},
			{Name: "KUIPER__BASIC__CONSOLELOG", Value: "true"},
		}))
		Expect(got.GetEKuiper().Ports).Should(ConsistOf([]corev1.ContainerPort{
			{Name: "fake-ekuiper", ContainerPort: 5678, Protocol: corev1.ProtocolTCP},
			{Name: "ekuiper", ContainerPort: 9081, Protocol: corev1.ProtocolTCP},
		}))
		Expect(got.GetEKuiper().ReadinessProbe).Should(Equal(&corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			FailureThreshold:    12,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "",
					Port: intstr.FromInt(9081),
				},
			},
		}))
		Expect(got.GetEKuiper().LivenessProbe).Should(Equal(&corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			FailureThreshold:    12,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "",
					Port: intstr.FromInt(9081),
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
			{Name: "ekuiper", Port: 9081, Protocol: corev1.ProtocolTCP, TargetPort: intstr.Parse("9081")},
			{Name: "fake-neuron", Port: 1234, Protocol: corev1.ProtocolTCP, TargetPort: intstr.Parse("1234")},
			{Name: "fake-ekuiper", Port: 5678, Protocol: corev1.ProtocolTCP, TargetPort: intstr.Parse("5678")},
		}))

		Expect(got.GetVolumeClaimTemplate().Name).Should(Equal(got.GetResName()))
		Expect(got.GetVolumeClaimTemplate().Namespace).Should(Equal(got.GetNamespace()))
		Expect(got.GetVolumeClaimTemplate().Annotations).Should(Equal(got.GetAnnotations()))
		Expect(got.GetServiceTemplate().Annotations).Should(Equal(got.GetAnnotations()))
	})
})
