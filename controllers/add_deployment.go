package controllers

import (
	"context"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/emqx/edge-operator/internal"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type addEkuiperDeployment struct{}

func (a addEkuiperDeployment) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.EKuiper) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add ekuiper Deployment")

	deploy := getDeployment(instance)
	if err := r.createOrUpdate(ctx, instance, &deploy, logger); err != nil {
		return &requeue{curError: err}
	}
	return nil
}

type addNeuronDeployment struct{}

func (a addNeuronDeployment) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.Neuron) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler", "add Neuron Deployment")

	deploy := getDeployment(instance)
	if err := r.createOrUpdate(ctx, instance, &deploy, logger); err != nil {
		return &requeue{curError: err}
	}
	return nil
}

type addNeuronExDeploy struct{}

func (a addNeuronExDeploy) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.NeuronEX) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add NeuronEx Deploy")

	deploy := getDeployment(instance)
	if err := r.createOrUpdate(ctx, instance, &deploy, logger); err != nil {
		return &requeue{curError: err}
	}
	return nil
}

func getDeployment(instance edgev1alpha1.EdgeInterface) appsv1.Deployment {
	podTemp := getPodTemplate(instance)

	deploy := appsv1.Deployment{
		ObjectMeta: internal.GetObjectMetadata(instance, instance.GetName()),
		Spec: appsv1.DeploymentSpec{
			Replicas: instance.GetReplicas(),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: podTemp.GetLabels(),
			},
			Template: podTemp,
		},
	}
	deploy.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("Deployment"))

	return deploy
}

func getPodTemplate(instance edgev1alpha1.EdgeInterface) corev1.PodTemplateSpec {
	pod := corev1.PodTemplateSpec{
		ObjectMeta: internal.GetObjectMetadata(instance, ""),
		Spec:       getPodSpec(instance),
	}

	return pod
}

func getPodSpec(instance edgev1alpha1.EdgeInterface) corev1.PodSpec {
	podSpec := &corev1.PodSpec{}
	edgePodSpec := instance.GetEdgePodSpec()
	structAssign(podSpec, &edgePodSpec)

	vols := getVolumeList(instance)
	for i := range vols {
		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name:         vols[i].name,
			VolumeSource: vols[i].volumeSource,
		})
	}

	switch instance.GetComponentType() {
	case edgev1alpha1.ComponentTypeNeuronEx:
		podSpec.Containers = []corev1.Container{
			getNeuronContainer(instance, vols),
			getEkuiperContainer(instance, vols),
		}
	case edgev1alpha1.ComponentTypeEKuiper:
		podSpec.Containers = []corev1.Container{
			getEkuiperContainer(instance, vols),
		}
	case edgev1alpha1.ComponentTypeNeuron:
		podSpec.Containers = []corev1.Container{
			getNeuronContainer(instance, vols),
		}
	default:
		panic("Unknown component " + instance.GetComponentType())
	}
	return *podSpec
}

func getNeuronContainer(ins edgev1alpha1.EdgeInterface, vols []volumeInfo) corev1.Container {
	container := ins.GetNeuron().DeepCopy()
	appendVolumeMount(container, mountToNeuron, vols)
	return *container
}

func getEkuiperContainer(ins edgev1alpha1.EdgeInterface, vols []volumeInfo) corev1.Container {
	container := ins.GetEKuiper().DeepCopy()
	appendVolumeMount(container, mountToEkuiper, vols)
	return *container
}

func appendVolumeMount(container *corev1.Container, mount mountTo, vols []volumeInfo) {
	for i := range vols {
		if attr, ok := vols[i].mounts[mount]; ok {
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      vols[i].name,
				MountPath: attr.path,
				ReadOnly:  attr.readOnly,
			})
		}
	}
}
