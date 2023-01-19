package controllers

import (
	"fmt"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("add deployment", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprint("add-deployment", +rand.Intn(10000)),
				Labels: map[string]string{
					"test": "e2e",
				},
			},
		}
		Expect(k8sClient.Create(ctx, namespace)).Should(Succeed())
	})

	DescribeTable("check deployment volumes",
		func(compType edgev1alpha1.ComponentType) {
			var ins edgev1alpha1.EdgeInterface
			switch compType {
			case edgev1alpha1.ComponentTypeEKuiper:
				ins = getEKuiper()
			case edgev1alpha1.ComponentTypeNeuron:
				ins = getNeuron()
			case edgev1alpha1.ComponentTypeNeuronEx:
				neuronEx := getNeuronEX()
				neuronEx.Spec.VolumeClaimTemplate = &corev1.PersistentVolumeClaimTemplate{
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("8Mi"),
							},
						},
					},
				}
				ins = neuronEx
				ins.Default()
			}

			ins.SetNamespace(namespace.Name)

			defer func() {
				Expect(k8sClient.Delete(ctx, ins)).Should(Succeed())
			}()

			Expect(k8sClient.Create(ctx, ins)).Should(Succeed())

			expectedVolumes := getVolumeList(ins)

			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ins.GetResName(),
					Namespace: ins.GetNamespace(),
				},
			}

			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
			}, timeout, interval).Should(Succeed())

			targetVolumes := deployment.Spec.Template.Spec.Volumes
			Expect(targetVolumes).Should(HaveLen(len(expectedVolumes)))

			isEqual := func(expected, target any) bool {
				if expected != nil && target != nil {
					return true
				}
				if expected == nil && target == nil {
					return true
				}
				return false
			}
			checkExisting := func(expected *volumeInfo) bool {
				for _, target := range targetVolumes {
					if target.Name == expected.name {
						if !isEqual(expected.volumeSource.EmptyDir, target.EmptyDir) {
							return false
						}
						if !isEqual(expected.volumeSource.ConfigMap, target.ConfigMap) {
							return false
						}
						if !isEqual(expected.volumeSource.PersistentVolumeClaim, target.PersistentVolumeClaim) {
							return false
						}
						return true
					}
				}
				return false
			}
			for _, expected := range expectedVolumes {
				Expect(checkExisting(&expected)).Should(BeTrue())
			}

			if compType == edgev1alpha1.ComponentTypeNeuronEx || compType == edgev1alpha1.ComponentTypeNeuron {
				neuron := ins.GetNeuron()

				isContain := false
				for _, container := range deployment.Spec.Template.Spec.Containers {
					if container.Name == neuron.Name {
						isContain = true
						Expect(container.Image).Should(Equal(neuron.Image))
						Expect(container.Ports).Should(ConsistOf([]corev1.ContainerPort{
							{Name: "neuron", ContainerPort: 7000, Protocol: corev1.ProtocolTCP},
						}))
						Expect(container.Env).Should(ConsistOf([]corev1.EnvVar{
							{Name: "LOG_CONSOLE", Value: "1"},
						}))

					}
				}
				Expect(isContain).Should(BeTrue())
			}

			if compType == edgev1alpha1.ComponentTypeNeuronEx || compType == edgev1alpha1.ComponentTypeEKuiper {
				eKuiper := ins.GetEKuiper()

				isContain := false
				for _, container := range deployment.Spec.Template.Spec.Containers {
					if container.Name == eKuiper.Name {
						isContain = true
						Expect(container.Name).Should(Equal(eKuiper.Name))
						Expect(container.Image).Should(Equal(eKuiper.Image))
						Expect(container.Env).Should(ConsistOf([]corev1.EnvVar{
							{Name: "KUIPER__BASIC__RESTPORT", Value: "9081"},
							{Name: "KUIPER__BASIC__IGNORECASE", Value: "false"},
							{Name: "KUIPER__BASIC__CONSOLELOG", Value: "true"},
						}))
						Expect(container.Ports).Should(ConsistOf([]corev1.ContainerPort{
							{Name: "ekuiper", ContainerPort: 9081, Protocol: corev1.ProtocolTCP},
						}))
					}
				}
				Expect(isContain).Should(BeTrue())
			}
		},
		Entry("neuronEX", edgev1alpha1.ComponentTypeNeuronEx),
		Entry("neuron", edgev1alpha1.ComponentTypeNeuron),
		Entry("ekuiper", edgev1alpha1.ComponentTypeEKuiper),
	)
})

var _ = Describe("update deployment", func() {
	var neuronEX *edgev1alpha1.NeuronEX
	var namespace *corev1.Namespace

	BeforeEach(func() {
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprint("update-deployment", +rand.Intn(10000)),
				Labels: map[string]string{
					"test": "e2e",
				},
			},
		}

		neuronEX = getNeuronEX()
		neuronEX.Namespace = namespace.Name

		Expect(k8sClient.Create(ctx, namespace)).Should(Succeed())
		Expect(k8sClient.Create(ctx, neuronEX)).Should(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, neuronEX)).Should(Succeed())
	})

	Context("update image", func() {
		BeforeEach(func() {
			newNeuronEX := neuronEX.DeepCopy()
			newNeuronEX.Spec.Neuron.Image = "emqx/neuron:latest"
			newNeuronEX.Spec.EKuiper.Image = "lfedge/ekuiper:latest-slim"
			Expect(k8sClient.Patch(ctx, newNeuronEX, client.MergeFrom(neuronEX))).Should(Succeed())
		})

		It("should get new image", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      neuronEX.GetResName(),
					Namespace: neuronEX.GetNamespace(),
				},
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
			}, timeout, interval).Should(Succeed())

			Expect(deployment.Spec.Template.Spec.Containers).Should(HaveLen(2))

			Expect([]string{
				deployment.Spec.Template.Spec.Containers[0].Image,
				deployment.Spec.Template.Spec.Containers[1].Image,
			}).Should(ConsistOf([]string{
				"emqx/neuron:latest",
				"lfedge/ekuiper:latest-slim",
			}))
		})
	})
})

/*var _ = Describe("update deployment", func() {
	var ekuiper = getEKuiper()

	BeforeEach(func() {
		Expect(k8sClient.Create(ctx, ekuiper.DeepCopy())).Should(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, ekuiper)).Should(Succeed())
	})

	Context("update deployment", func() {
		deployment := &appsv1.Deployment{}
		var err error
		BeforeEach(func() {
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      ekuiper.GetResName(),
					Namespace: ekuiper.GetNamespace(),
				}, deployment)
			}, timeout, interval).Should(Succeed())

			newEKuiper := ekuiper.DeepCopy()
			newEKuiper.Annotations["update"] = "test"
			err = k8sClient.Patch(ctx, newEKuiper, client.MergeFrom(ekuiper))
		})

		It("should succeed", func() {
			Expect(err).Should(Succeed())
		})

		When("deploy has been updated", func() {
			Context("check annotation", func() {
				It("should have new annotation", func() {
					Eventually(func() map[string]string {
						err = k8sClient.Get(ctx, types.NamespacedName{
							Name:      ekuiper.GetResName(),
							Namespace: ekuiper.GetNamespace(),
						}, deployment)
						Expect(err).Should(Succeed())

						return deployment.Annotations
					}, timeout, interval).Should(HaveKey("update"))
				})
			})
		})
	})
})
*/
