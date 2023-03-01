package controllers

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/emqx/edge-operator/internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("add secret", func() {
	neuronEX := getNeuronEX()
	neuron := getNeuron()
	ekuiper := getEKuiper()

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

	DescribeTable("should create default secret",
		func(ins edgev1alpha1.EdgeInterface) {
			Eventually(func() []corev1.Secret {
				secrets := &corev1.SecretList{}
				_ = k8sClient.List(ctx, secrets, client.InNamespace(ins.GetNamespace()), client.MatchingLabels(ins.GetLabels()))
				return secrets.Items
			}, timeout, interval).Should(HaveLen(1))
		},
		Entry("neuronEX", neuronEX),
		Entry("neuron", neuron),
		Entry("ekuiper", ekuiper),
	)
})

var _ = Describe("add secret", func() {
	neuronEX := getNeuronEX()
	neuron := getNeuron()
	ekuiper := getEKuiper()
	publicKeys := []edgev1alpha1.PublicKey{
		{
			Name: "sample-file",
			Data: []byte("base64encodingData"),
		},
		{
			Name: "sample-file2",
			Data: []byte("base64encodingData"),
		},
	}

	BeforeEach(func() {
		neuronEX.Spec.PublicKeys = publicKeys
		neuron.Spec.PublicKeys = publicKeys
		ekuiper.Spec.PublicKeys = publicKeys
		Expect(k8sClient.Create(ctx, neuronEX.DeepCopy())).Should(Succeed())
		Expect(k8sClient.Create(ctx, neuron.DeepCopy())).Should(Succeed())
		Expect(k8sClient.Create(ctx, ekuiper.DeepCopy())).Should(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, neuronEX)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, neuron)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, ekuiper)).Should(Succeed())
	})

	DescribeTable("auth file secret should been created",
		func(ins edgev1alpha1.EdgeInterface, publicKeys []edgev1alpha1.PublicKey) {
			secrets := &corev1.SecretList{}

			Eventually(func() []corev1.Secret {
				_ = k8sClient.List(ctx, secrets, client.InNamespace(ins.GetNamespace()), client.MatchingLabels(ins.GetLabels()))
				return secrets.Items
			}, timeout, interval).Should(HaveLen(1))

			Eventually(func() map[string][]byte {
				_ = k8sClient.List(ctx, secrets, client.InNamespace(ins.GetNamespace()), client.MatchingLabels(ins.GetLabels()))
				return secrets.Items[0].Data
			}, timeout, interval).Should(HaveLen(len(publicKeys)))

			secret := secrets.Items[0]
			Expect(secret.Name).Should(Equal(internal.GetResNameOnPanic(ins, publicKey)))
			Expect(secret.Labels).Should(Equal(ins.GetLabels()))
			Expect(secret.Annotations).Should(HaveKeyWithValue("foo", "bar"))

			for i := range publicKeys {
				Expect(secret.Data).Should(
					HaveKeyWithValue(publicKeys[i].Name, publicKeys[i].Data))
			}
		},
		Entry("neuronEX", neuronEX, publicKeys),
		Entry("neuron", neuron, publicKeys),
		Entry("ekuiper", ekuiper, publicKeys),
	)
})
