package controllers

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

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
		ObjectMeta: internal.GetObjectMetadata(instance, instance.GetResName()),
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
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
			getNeuronContainer(instance, vols...),
			getEkuiperContainer(instance, vols...),
			getEkuiperToolContainer(instance, vols...),
		}
	case edgev1alpha1.ComponentTypeEKuiper:
		podSpec.Containers = []corev1.Container{
			getEkuiperContainer(instance, vols...),
		}
	case edgev1alpha1.ComponentTypeNeuron:
		podSpec.Containers = []corev1.Container{
			getNeuronContainer(instance, vols...),
		}
	default:
		panic("unknown component " + instance.GetComponentType())
	}
	return *podSpec
}

func getNeuronContainer(ins edgev1alpha1.EdgeInterface, vols ...volumeInfo) corev1.Container {
	container := ins.GetNeuron().DeepCopy()
	for i := range vols {
		for _, m := range vols[i].mountTo {
			if m == mountToNeuron {
				container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
					Name:      vols[i].name,
					MountPath: vols[i].mountPath,
				})
			}
		}
	}
	return *container
}

func getEkuiperContainer(ins edgev1alpha1.EdgeInterface, vols ...volumeInfo) corev1.Container {
	container := ins.GetEKuiper().DeepCopy()
	for i := range vols {
		for _, m := range vols[i].mountTo {
			if m == mountToEkuiper {
				container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
					Name:      vols[i].name,
					MountPath: vols[i].mountPath,
				})
			}
		}
	}
	return *container
}

func getEkuiperToolContainer(ins edgev1alpha1.EdgeInterface, vols ...volumeInfo) corev1.Container {
	compile := regexp.MustCompile(`[0-9]+(\.[0-9]+)?(\.[0-9]+)?(-(alpha|beta|rc)\.[0-9]+)?`)

	i := strings.Split(ins.GetEKuiper().Image, ":")
	registry := filepath.Dir(i[0])
	version := "latest"
	if compile.MatchString(i[1]) {
		version = compile.FindString(i[1])
	}

	container := corev1.Container{
		Name:            "ekuiper-tool",
		Image:           registry + "/ekuiper-kubernetes-tool:" + version,
		ImagePullPolicy: ins.GetEKuiper().ImagePullPolicy,
	}
	for i := range vols {
		for _, m := range vols[i].mountTo {
			if m == mountToEkuiperTool {
				container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
					Name:      vols[i].name,
					MountPath: vols[i].mountPath,
					ReadOnly:  true,
				})
			}
		}
	}
	return container
}
