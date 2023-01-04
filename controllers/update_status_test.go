package controllers

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("basic test", func() {
	DescribeTable("check custom resources status",
		func(compType edgev1alpha1.ComponentType) {
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

			By("create cr")
			Expect(k8sClient.Create(ctx, ins)).Should(Succeed())

			By("update deployment readyReplicas for target controller reconcile")
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ins.GetResName(),
					Namespace: ins.GetNamespace(),
				},
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
			}, timeout, interval).Should(Succeed())

			By("patch status for deployment")
			patchDeploy := deployment.DeepCopy()
			patchDeploy.Status.Replicas = 1
			patchDeploy.Status.ReadyReplicas = 1
			Expect(k8sClient.Status().Patch(ctx, patchDeploy, client.StrategicMergeFrom(deployment))).Should(Succeed())

			By("check cr status")
			Eventually(func() edgev1alpha1.CRPhase {
				got := deepCopyEdgeEdgeInterface(ins)
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(ins), got)
				return got.GetStatus().Phase
			}, timeout, interval).Should(Equal(edgev1alpha1.CRReady))
		},
		Entry("neuronEX", edgev1alpha1.ComponentTypeNeuronEx),
		Entry("neuron", edgev1alpha1.ComponentTypeNeuron),
		Entry("ekuiper", edgev1alpha1.ComponentTypeEKuiper),
	)
})
