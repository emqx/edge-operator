package controllers

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/emqx/edge-operator/internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("check deployment when volume template not set", func() {
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

	DescribeTable("check deployment volumes",
		func(ins edgev1alpha1.EdgeInterface, expected []corev1.Volume) {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ins.GetResName(),
					Namespace: ins.GetNamespace(),
				},
			}
			Eventually(func() []corev1.Volume {
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
				return deployment.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(ConsistOf(expected))
		},
		Entry("neuronEX", neuronEX, []corev1.Volume{
			{Name: "shared-tmp", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			{Name: "neuron-data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			{Name: "ekuiper-data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			{Name: "ekuiper-plugins", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			{Name: "ekuiper-tool-config", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: internal.GetResNameOnPanic(neuronEX, ekuiperToolConfig)},
				DefaultMode:          &[]int32{corev1.ConfigMapVolumeSourceDefaultMode}[0],
			}}}}),
		Entry("neuron", neuron, []corev1.Volume{
			{Name: "neuron-data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
		}),
		Entry("ekuiper", ekuiper, []corev1.Volume{
			{Name: "ekuiper-data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			{Name: "ekuiper-plugins", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
		}),
	)
})

var _ = Describe("check deployment when volume template set", func() {
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

	DescribeTable("check deployment",
		func(ins edgev1alpha1.EdgeInterface, expected []corev1.Volume) {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ins.GetResName(),
					Namespace: ins.GetNamespace(),
				},
			}
			Eventually(func() []corev1.Volume {
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
				return deployment.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(ConsistOf(expected))
		},
		Entry("neuronEX", addVolumeTemplate(neuronEX), []corev1.Volume{
			{Name: "shared-tmp", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			{Name: "neuron-data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: addVolumeTemplate(neuronEX).GetVolumeClaimTemplate().Name + "-neuron-data"}}},
			{Name: "ekuiper-data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: addVolumeTemplate(neuronEX).GetVolumeClaimTemplate().Name + "-ekuiper-data"}}},
			{Name: "ekuiper-plugins", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: addVolumeTemplate(neuronEX).GetVolumeClaimTemplate().Name + "-ekuiper-plugins"}}},
			{Name: "ekuiper-tool-config", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: internal.GetResNameOnPanic(neuronEX, ekuiperToolConfig)},
				DefaultMode:          &[]int32{corev1.ConfigMapVolumeSourceDefaultMode}[0],
			}}}}),
		Entry("neuron", addVolumeTemplate(neuron), []corev1.Volume{
			{Name: "neuron-data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: addVolumeTemplate(neuron).GetVolumeClaimTemplate().Name + "-neuron-data"}}},
		}),
		Entry("ekuiper", addVolumeTemplate(ekuiper), []corev1.Volume{
			{Name: "ekuiper-data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: addVolumeTemplate(ekuiper).GetVolumeClaimTemplate().Name + "-ekuiper-data"}}},
			{Name: "ekuiper-plugins", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: addVolumeTemplate(ekuiper).GetVolumeClaimTemplate().Name + "-ekuiper-plugins"}}},
		}),
	)
})

var _ = Describe("check deployment and containers", func() {
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

	DescribeTable("check deployment",
		func(ins edgev1alpha1.EdgeInterface, expectedContainerCount int) {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ins.GetResName(),
					Namespace: ins.GetNamespace(),
				},
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
			}, timeout, interval).Should(Succeed())

			// metadata
			Expect(deployment.ObjectMeta.Labels).Should(Equal(ins.GetLabels()))
		},
		Entry("neuronEX", neuronEX, 3),
		Entry("neuron", neuron, 1),
		Entry("ekuiper", ekuiper, 1),
	)

	It("check neuronEX container", func() {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      neuronEX.GetResName(),
				Namespace: neuronEX.GetNamespace(),
			},
		}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
		}, timeout, interval).Should(Succeed())

		// neuron container
		Expect(deployment.Spec.Template.Spec.Containers[0].Env).Should(ConsistOf([]corev1.EnvVar{
			{Name: "LOG_CONSOLE", Value: "1"},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[0].Name).Should(Equal(neuronEX.Spec.Neuron.Name))
		Expect(deployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(neuronEX.Spec.Neuron.Image))
		Expect(deployment.Spec.Template.Spec.Containers[0].Ports).Should(ConsistOf([]corev1.ContainerPort{
			{Name: "neuron", ContainerPort: 7000, Protocol: corev1.ProtocolTCP},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts).Should(ConsistOf([]corev1.VolumeMount{
			{Name: "shared-tmp", MountPath: "/tmp"},
			{Name: "neuron-data", MountPath: "/opt/neuron/persistence"},
		}))
		// ekuiper container
		Expect(deployment.Spec.Template.Spec.Containers[1].Name).Should(Equal(neuronEX.Spec.EKuiper.Name))
		Expect(deployment.Spec.Template.Spec.Containers[1].Image).Should(Equal(neuronEX.Spec.EKuiper.Image))
		Expect(deployment.Spec.Template.Spec.Containers[1].Env).Should(ConsistOf([]corev1.EnvVar{
			{Name: "KUIPER__BASIC__RESTPORT", Value: "9081"},
			{Name: "KUIPER__BASIC__IGNORECASE", Value: "false"},
			{Name: "KUIPER__BASIC__CONSOLELOG", Value: "true"},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[1].Ports).Should(ConsistOf([]corev1.ContainerPort{
			{Name: "ekuiper", ContainerPort: 9081, Protocol: corev1.ProtocolTCP},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[1].VolumeMounts).Should(ConsistOf([]corev1.VolumeMount{
			{Name: "shared-tmp", MountPath: "/tmp"},
			{Name: "ekuiper-data", MountPath: "/kuiper/data"},
			{Name: "ekuiper-plugins", MountPath: "/kuiper/plugins/portable"},
		}))
		// ekuiper tool container
		Expect(deployment.Spec.Template.Spec.Containers[2].Name).Should(Equal("ekuiper-tool"))
		Expect(deployment.Spec.Template.Spec.Containers[2].Image).Should(Equal("lfedge/ekuiper-kubernetes-tool:1.8"))
		Expect(deployment.Spec.Template.Spec.Containers[2].VolumeMounts).Should(ConsistOf([]corev1.VolumeMount{
			{Name: "ekuiper-tool-config", MountPath: "/kuiper-kubernetes-tool/sample", ReadOnly: true},
		}))
	})

	It("check neuron container", func() {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      neuron.GetResName(),
				Namespace: neuron.GetNamespace(),
			},
		}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
		}, timeout, interval).Should(Succeed())
		// neuron container
		Expect(deployment.Spec.Template.Spec.Containers[0].Env).Should(ConsistOf([]corev1.EnvVar{
			{Name: "LOG_CONSOLE", Value: "1"},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[0].Name).Should(Equal(neuron.Spec.Neuron.Name))
		Expect(deployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(neuron.Spec.Neuron.Image))
		Expect(deployment.Spec.Template.Spec.Containers[0].Ports).Should(ConsistOf([]corev1.ContainerPort{
			{Name: "neuron", ContainerPort: 7000, Protocol: corev1.ProtocolTCP},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts).Should(ConsistOf([]corev1.VolumeMount{
			{Name: "neuron-data", MountPath: "/opt/neuron/persistence"},
		}))
	})

	It("check ekuiper container", func() {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ekuiper.GetResName(),
				Namespace: ekuiper.GetNamespace(),
			},
		}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
		}, timeout, interval).Should(Succeed())
		// ekuiper container
		Expect(deployment.Spec.Template.Spec.Containers[0].Name).Should(Equal(ekuiper.Spec.EKuiper.Name))
		Expect(deployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(ekuiper.Spec.EKuiper.Image))
		Expect(deployment.Spec.Template.Spec.Containers[0].Env).Should(ConsistOf([]corev1.EnvVar{
			{Name: "KUIPER__BASIC__RESTPORT", Value: "9081"},
			{Name: "KUIPER__BASIC__IGNORECASE", Value: "false"},
			{Name: "KUIPER__BASIC__CONSOLELOG", Value: "true"},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[0].Ports).Should(ConsistOf([]corev1.ContainerPort{
			{Name: "ekuiper", ContainerPort: 9081, Protocol: corev1.ProtocolTCP},
		}))
		Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts).Should(ConsistOf([]corev1.VolumeMount{
			{Name: "ekuiper-data", MountPath: "/kuiper/data"},
			{Name: "ekuiper-plugins", MountPath: "/kuiper/plugins/portable"},
		}))
	})

	Describe("check update deployment", func() {
		JustBeforeEach(func() {
			newNeuronEX := neuronEX.DeepCopy()
			newNeuronEX.Spec.Neuron.Image = "emqx/neuron:latest"
			newNeuronEX.Spec.EKuiper.Image = "lfedge/ekuiper:latest-slim"
			Expect(k8sClient.Patch(ctx, newNeuronEX, client.MergeFrom(neuronEX))).Should(Succeed())

			newNeuron := neuron.DeepCopy()
			newNeuron.Spec.Neuron.Image = "emqx/neuron:latest"
			Expect(k8sClient.Patch(ctx, newNeuron, client.MergeFrom(neuron))).Should(Succeed())

			newEKuiper := ekuiper.DeepCopy()
			newEKuiper.Spec.EKuiper.Image = "lfedge/ekuiper:latest-slim"
			Expect(k8sClient.Patch(ctx, newEKuiper, client.MergeFrom(ekuiper))).Should(Succeed())
		})

		JustAfterEach(func() {
			pvcs := &corev1.PersistentVolumeClaimList{}
			Expect(k8sClient.List(ctx, pvcs, client.InNamespace("default"))).Should(Succeed())
			for _, pvc := range pvcs.Items {
				p := pvc.DeepCopy()
				p.SetFinalizers([]string{})
				Expect(k8sClient.Update(ctx, p)).Should(Succeed())
				Expect(k8sClient.Delete(ctx, p)).Should(Succeed())
			}
		})

		It("check neuron and ekuiper and ekuiper tool container", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      neuronEX.GetResName(),
					Namespace: neuronEX.GetNamespace(),
				},
			}
			Eventually(func() []string {
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
				return []string{
					deployment.Spec.Template.Spec.Containers[0].Image,
					deployment.Spec.Template.Spec.Containers[1].Image,
					deployment.Spec.Template.Spec.Containers[2].Image,
				}
			}, timeout, interval).Should(ConsistOf([]string{
				"emqx/neuron:latest",
				"lfedge/ekuiper:latest-slim",
				"lfedge/ekuiper-kubernetes-tool:latest",
			}))
		})

		It("check neuron container", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      neuron.GetResName(),
					Namespace: neuron.GetNamespace(),
				},
			}
			Eventually(func() string {
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
				return deployment.Spec.Template.Spec.Containers[0].Image
			}, timeout, interval).Should(Equal("emqx/neuron:latest"))
		})

		It("check ekuiper container", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ekuiper.GetResName(),
					Namespace: ekuiper.GetNamespace(),
				},
			}
			Eventually(func() string {
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
				return deployment.Spec.Template.Spec.Containers[0].Image
			}, timeout, interval).Should(Equal("lfedge/ekuiper:latest-slim"))
		})
	})
})
