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

var _ = Describe("EKuiper validate webhook", func() {
	var ins *EKuiper

	BeforeEach(func() {
		ins = new(EKuiper)
		ins = &EKuiper{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ekuiper",
				Namespace: "default",
			},
		}
	})

	AfterEach(func() {
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ins), ins); err == nil {
			Expect(k8sClient.Delete(ctx, ins.DeepCopy())).Should(Succeed())
		}
	})

	It("should block if ekuiper name is empty", func() {
		ins.Spec.EKuiper.Name = ""
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(MatchError(matchError(ins, "ekuiper container name is empty")))
	})

	It("should block if ekuiper image is empty", func() {
		ins.Spec.EKuiper.Name = "ekuiper"
		ins.Spec.EKuiper.Image = ""
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(MatchError(matchError(ins, "ekuiper container image is empty")))
	})
	It("should block create if ekuiper image is not slim or slim-python", func() {
		ins.Spec.EKuiper.Name = "ekuiper"
		ins.Spec.EKuiper.Image = "lfedge/ekuiper:1.8"
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(MatchError(matchError(ins, "ekuiper container image must be slim or slim-python")))
		ins.Spec.EKuiper.Image = "lfedge/ekuiper:1.8-slim"
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(Succeed())
	})

	It("should block update if ekuiper name is empty", func() {
		ins.Spec.EKuiper.Name = "ekuiper"
		ins.Spec.EKuiper.Image = "lfedge/ekuiper:1.8-slim-python"
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(Succeed())

		new := ins.DeepCopy()
		new.Spec.EKuiper.Name = ""
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(MatchError(matchError(ins, "ekuiper container name is empty")))

		new.Spec.EKuiper.Name = "ekuiper"
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(Succeed())
	})

	It("should block update if ekuiper image is empty", func() {
		ins.Spec.EKuiper.Name = "ekuiper"
		ins.Spec.EKuiper.Image = "lfedge/ekuiper:1.8-slim-python"
		Expect(k8sClient.Create(ctx, ins.DeepCopy())).Should(Succeed())

		new := ins.DeepCopy()
		new.Spec.EKuiper.Image = ""
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(MatchError(matchError(ins, "ekuiper container image is empty")))

		new.Spec.EKuiper.Image = "lfedge/ekuiper:latest-slim-python"
		Expect(k8sClient.Patch(ctx, new.DeepCopy(), client.MergeFrom(ins.DeepCopy()))).Should(Succeed())
	})

	It("should block update if ekuiper image is slim-python or not slim-python", func() {
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
