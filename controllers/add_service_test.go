package controllers

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("check service", Label("service"), func() {
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

	DescribeTable("service should not created",
		func(ins edgev1alpha1.EdgeInterface) {
			list := &corev1.ServiceList{}
			Expect(k8sClient.List(ctx, list, client.InNamespace(ins.GetNamespace()), client.MatchingLabels(ins.GetLabels()))).Should(Succeed())
			Expect(list.Items).Should(BeEmpty())
		},
		Entry("neuronEX", getNeuronEX()),
		Entry("neuron", neuron),
		Entry("ekuiper", ekuiper),
	)

	Describe("update service", func() {
		JustBeforeEach(func() {
			Expect(k8sClient.Patch(ctx, addServiceTemplate(neuronEX), client.MergeFrom(neuronEX))).Should(Succeed())
			Expect(k8sClient.Patch(ctx, addServiceTemplate(neuron), client.MergeFrom(neuron))).Should(Succeed())
			Expect(k8sClient.Patch(ctx, addServiceTemplate(ekuiper), client.MergeFrom(ekuiper))).Should(Succeed())
		})

		DescribeTable("service should created",
			func(ins edgev1alpha1.EdgeInterface, expected []corev1.ServicePort) {
				service := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      ins.GetServiceTemplate().Name,
						Namespace: ins.GetNamespace(),
					},
				}
				Eventually(func() error {
					return k8sClient.Get(ctx, client.ObjectKeyFromObject(service), service)
				}, timeout, interval).Should(Succeed())
				Expect(service.Annotations).Should(HaveKeyWithValue("foo", "bar"))
				Expect(service.Annotations).Should(HaveKeyWithValue("e2e/test", "serviceTemplate"))
				Expect(service.Spec.Ports).Should(ConsistOf(expected))
			},
			Entry("neuronEX", addServiceTemplate(neuronEX), []corev1.ServicePort{
				{Name: "neuron", Port: 7000, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(7000)},
				{Name: "ekuiper", Port: 9081, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(9081)},
			}),
			Entry("neuron", addServiceTemplate(neuron), []corev1.ServicePort{
				{Name: "neuron", Port: 7000, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(7000)},
			}),
			Entry("ekuiper", addServiceTemplate(ekuiper), []corev1.ServicePort{
				{Name: "ekuiper", Port: 9081, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(9081)},
			}),
		)
	})
})
