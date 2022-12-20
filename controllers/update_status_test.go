package controllers

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("basic test", Label("basic"), func() {
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

	DescribeTable("check custom resources status",
		func(ins edgev1alpha1.EdgeInterface) {
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
				got := deepCopyEdgeEdgeInterface(ins)
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(ins), got)
				return got.GetStatus().Phase
			}, timeout, interval).ShouldNot(Equal(corev1.PodRunning))

			// Update pod status
			pod.Status.Phase = corev1.PodRunning
			Expect(k8sClient.Status().Update(ctx, pod)).Should(Succeed())

			// Update deployment for target controller reconcile
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ins.GetResName(),
					Namespace: ins.GetNamespace(),
				},
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
			}, timeout, interval).Should(Succeed())
			deployment.Annotations["target-reconcile"] = "true"
			Expect(k8sClient.Update(ctx, deployment)).Should(Succeed())

			// Check status
			Eventually(func() corev1.PodPhase {
				got := deepCopyEdgeEdgeInterface(ins)
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(ins), got)
				return got.GetStatus().Phase
			}, timeout, interval).Should(Equal(corev1.PodRunning))
		},
		Entry("neuronEX", neuronEX),
		Entry("neuron", neuron),
		Entry("ekuiper", ekuiper),
	)
})
