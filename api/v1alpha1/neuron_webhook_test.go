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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
