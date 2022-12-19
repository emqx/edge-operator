package controllers

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("check pvc when volume template not set", func() {
	var neuronEX *edgev1alpha1.NeuronEX = getNeuronEX()
	var neuron *edgev1alpha1.Neuron = getNeuron()
	var ekuiper *edgev1alpha1.EKuiper = getEKuiper()

	BeforeEach(func() {
		Expect(k8sClient.Create(ctx, neuronEX.DeepCopy())).Should(Succeed())
		Expect(k8sClient.Create(ctx, neuron.DeepCopy())).Should(Succeed())
		Expect(k8sClient.Create(ctx, ekuiper.DeepCopy())).Should(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, neuronEX)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, neuron)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, ekuiper)).Should(Succeed())
	})

	DescribeTable("pvc should not created",
		func(ins edgev1alpha1.EdgeInterface) {
			list := &corev1.PersistentVolumeClaimList{}
			Expect(k8sClient.List(ctx, list, client.InNamespace(ins.GetNamespace()), client.MatchingLabels(ins.GetLabels()))).Should(Succeed())
			Expect(list.Items).Should(BeEmpty())
		},
		Entry("neuronEX", neuronEX),
		Entry("neuron", neuron),
		Entry("ekuiper", ekuiper),
	)
})

var _ = Describe("check pvc when volume template set", func() {
	var neuronEX *edgev1alpha1.NeuronEX = getNeuronEX()
	var neuron *edgev1alpha1.Neuron = getNeuron()
	var ekuiper *edgev1alpha1.EKuiper = getEKuiper()

	BeforeEach(func() {
		Expect(k8sClient.Create(ctx, addVolumeTemplate(neuronEX))).Should(Succeed())
		Expect(k8sClient.Create(ctx, addVolumeTemplate(neuron))).Should(Succeed())
		Expect(k8sClient.Create(ctx, addVolumeTemplate(ekuiper))).Should(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, neuronEX)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, neuron)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, ekuiper)).Should(Succeed())

		pvcs := &corev1.PersistentVolumeClaimList{}
		Expect(k8sClient.List(ctx, pvcs, client.InNamespace("default"))).Should(Succeed())
		for _, pvc := range pvcs.Items {
			p := pvc.DeepCopy()
			p.SetFinalizers([]string{})
			Expect(k8sClient.Update(ctx, p)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, p)).Should(Succeed())
		}
	})

	DescribeTable("pvc should created",
		func(ins edgev1alpha1.EdgeInterface, expected []string) {
			pvcList := &corev1.PersistentVolumeClaimList{}

			Eventually(func() []corev1.PersistentVolumeClaim {
				_ = k8sClient.List(ctx, pvcList, client.InNamespace(ins.GetNamespace()), client.MatchingLabels(ins.GetLabels()))
				return pvcList.Items
			}, timeout, interval).Should(HaveLen(len(expected)))

			for _, pvc := range pvcList.Items {
				Expect(pvc.Name).Should(BeElementOf(expected))
				Expect(pvc.Labels).Should(Equal(ins.GetLabels()))
				Expect(pvc.Annotations).Should(HaveKeyWithValue("foo", "bar"))

				Expect(pvc.Spec.AccessModes).Should(ConsistOf([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}))
			}
		},
		Entry("neuronEX", addVolumeTemplate(neuronEX), []string{
			neuronEX.GetResName() + "-neuron-data",
			neuronEX.GetResName() + "-ekuiper-data",
			neuronEX.GetResName() + "-ekuiper-plugins",
		}),
		Entry("neuron", addVolumeTemplate(neuron), []string{
			neuron.GetResName() + "-neuron-data",
		}),
		Entry("ekuiper", addVolumeTemplate(ekuiper), []string{
			ekuiper.GetResName() + "-ekuiper-data",
			ekuiper.GetResName() + "-ekuiper-plugins",
		}),
	)
})
