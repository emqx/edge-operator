package controllers

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/emqx/edge-operator/internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("basic test", Label("basic"), func() {
	var neuronEX *edgev1alpha1.NeuronEX = getNeuronEX()

	BeforeEach(func() {
		Expect(k8sClient.Create(ctx, neuronEX.DeepCopy())).Should(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, neuronEX)).Should(Succeed())
	})

	It("should create configMap", func() {
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      internal.GetResNameOnPanic(neuronEX, ekuiperToolConfig),
				Namespace: neuronEX.Namespace,
			},
		}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKeyFromObject(configMap), configMap)
		}, timeout, interval).Should(Succeed())

		// metadata
		Expect(configMap.ObjectMeta.Labels).Should(Equal(neuronEX.Labels))
		Expect(configMap.ObjectMeta.Annotations).Should(HaveKeyWithValue("foo", "bar"))
		// data
		Expect(configMap.Data).Should(HaveKey("neuronStream.json"))
	})
})
