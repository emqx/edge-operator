package controllers

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("check pvc when volume template set", func() {
	deletePVC := func() {
		pvcs := &corev1.PersistentVolumeClaimList{}
		Expect(k8sClient.List(ctx, pvcs, client.InNamespace("default"))).Should(Succeed())
		for _, pvc := range pvcs.Items {
			p := pvc.DeepCopy()
			p.SetFinalizers([]string{})
			Expect(k8sClient.Update(ctx, p)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, p)).Should(Succeed())
		}
	}

	BeforeEach(func() {
		deletePVC()
	})

	AfterEach(func() {
		deletePVC()
	})

	DescribeTable("pvc should created",
		func(compType edgev1alpha1.ComponentType, expected []string) {
			var ins edgev1alpha1.EdgeInterface
			switch compType {
			case edgev1alpha1.ComponentTypeEKuiper:
				ins = getEKuiper()
			case edgev1alpha1.ComponentTypeNeuron:
				ins = getNeuron()
			case edgev1alpha1.ComponentTypeNeuronEx:
				ins = getNeuronEX()
			}

			defer func() {
				Expect(k8sClient.Delete(ctx, ins)).Should(Succeed())
			}()

			By("create cr with no pvc")
			Expect(k8sClient.Create(ctx, ins)).Should(Succeed())

			pvcList := &corev1.PersistentVolumeClaimList{}
			Eventually(func() []corev1.PersistentVolumeClaim {
				err := k8sClient.List(ctx, pvcList, client.InNamespace(ins.GetNamespace()), client.MatchingLabels(ins.GetLabels()))
				Expect(err).Should(Succeed())
				return pvcList.Items
			}, timeout, interval).Should(BeEmpty())

			By("update cr with pvc")
			insWithVolumes := addVolumeTemplate(ins)
			Expect(k8sClient.Patch(ctx, insWithVolumes, client.MergeFrom(ins))).Should(Succeed())

			pvcList = &corev1.PersistentVolumeClaimList{}
			Eventually(func() []corev1.PersistentVolumeClaim {
				err := k8sClient.List(ctx, pvcList, client.InNamespace(ins.GetNamespace()), client.MatchingLabels(ins.GetLabels()))
				Expect(err).Should(Succeed())
				return pvcList.Items
			}, timeout, interval).Should(HaveLen(len(expected)))

			for _, pvc := range pvcList.Items {
				Expect(pvc.Name).Should(BeElementOf(expected))
				Expect(pvc.Labels).Should(Equal(ins.GetLabels()))
				Expect(pvc.Annotations).Should(HaveKeyWithValue("foo", "bar"))

				Expect(pvc.Spec.AccessModes).Should(ConsistOf([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}))
			}
		},
		Entry("neuronEX", edgev1alpha1.ComponentTypeNeuronEx, []string{
			"neuronex-neuron-data",
			"neuronex-ekuiper",
		}),
		Entry("neuron", edgev1alpha1.ComponentTypeNeuron, []string{
			"neuron-neuron-data",
		}),
		Entry("ekuiper", edgev1alpha1.ComponentTypeEKuiper, []string{
			"ekuiper-ekuiper",
		}),
	)
})
