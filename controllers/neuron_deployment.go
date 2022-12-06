package controllers

import (
	"context"
	"strings"

	emperror "emperror.dev/errors"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type neuronDeploy struct {
	pvcSubReconciler addNeuronPVC
}

func newNeuronDeploy() neuronDeploy {
	return neuronDeploy{
		pvcSubReconciler: addNeuronPVC{},
	}
}

func (sub neuronDeploy) reconcile(ctx context.Context, r *NeuronReconciler,
	instance *edgev1alpha1.Neuron) *requeue {
	if err := sub.pvcSubReconciler.reconcile(ctx, r, instance.Spec.VolumeClaimTemplate); err != nil {
		return err
	}

	deploy := sub.getDeployment(instance)
	sub.updateStorage(deploy, sub.pvcSubReconciler.GetNeuronClaimList(instance.Spec.VolumeClaimTemplate),
		instance.Spec.Neuron.Name)

	if err := createOrUpdateNeuron(ctx, r, instance, deploy); err != nil {
		return &requeue{curError: emperror.Wrap(err, "failed to create or update deployment")}
	}

	return nil
}

func (sub neuronDeploy) getDeployment(instance *edgev1alpha1.Neuron) *appsv1.Deployment {
	labels := instance.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	delete(labels, "kubectl.kubernetes.io/last-applied-configuration")

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Annotations: instance.Annotations,
			Labels:      labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: instance.Annotations,
				},
				Spec: sub.getPodSpec(instance),
			},
		},
	}
	deploy.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("Deployment"))
	return deploy
}

func (sub neuronDeploy) getPodSpec(instance *edgev1alpha1.Neuron) corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{
			*sub.getNeuronContainer(instance.Spec.Neuron.DeepCopy()),
		},
		Volumes:                       instance.Spec.Volumes,
		InitContainers:                instance.Spec.InitContainers,
		EphemeralContainers:           instance.Spec.EphemeralContainers,
		RestartPolicy:                 instance.Spec.RestartPolicy,
		TerminationGracePeriodSeconds: instance.Spec.TerminationGracePeriodSeconds,
		ActiveDeadlineSeconds:         instance.Spec.ActiveDeadlineSeconds,
		DNSPolicy:                     instance.Spec.DNSPolicy,
		NodeSelector:                  instance.Spec.NodeSelector,
		ServiceAccountName:            instance.Spec.ServiceAccountName,
		DeprecatedServiceAccount:      instance.Spec.DeprecatedServiceAccount,
		AutomountServiceAccountToken:  instance.Spec.AutomountServiceAccountToken,
		NodeName:                      instance.Spec.NodeName,
		HostNetwork:                   instance.Spec.HostNetwork,
		HostPID:                       instance.Spec.HostPID,
		HostIPC:                       instance.Spec.HostIPC,
		ShareProcessNamespace:         instance.Spec.ShareProcessNamespace,
		SecurityContext:               instance.Spec.PodSecurityContext,
		ImagePullSecrets:              instance.Spec.ImagePullSecrets,
		Hostname:                      instance.Spec.Hostname,
		Subdomain:                     instance.Spec.Subdomain,
		Affinity:                      instance.Spec.Affinity,
		SchedulerName:                 instance.Spec.SchedulerName,
		Tolerations:                   instance.Spec.Tolerations,
		HostAliases:                   instance.Spec.HostAliases,
		PriorityClassName:             instance.Spec.PriorityClassName,
		Priority:                      instance.Spec.Priority,
		DNSConfig:                     instance.Spec.DNSConfig,
		ReadinessGates:                instance.Spec.ReadinessGates,
		RuntimeClassName:              instance.Spec.RuntimeClassName,
		EnableServiceLinks:            instance.Spec.EnableServiceLinks,
		PreemptionPolicy:              instance.Spec.PreemptionPolicy,
		Overhead:                      instance.Spec.Overhead,
		TopologySpreadConstraints:     instance.Spec.TopologySpreadConstraints,
		SetHostnameAsFQDN:             instance.Spec.SetHostnameAsFQDN,
		OS:                            instance.Spec.OS,
		HostUsers:                     instance.Spec.HostUsers,
	}
}

func (sub neuronDeploy) getNeuronContainer(neuron *corev1.Container) *corev1.Container {
	return neuron
}

func (sub neuronDeploy) updateStorage(deploy *appsv1.Deployment, claimList []*corev1.PersistentVolumeClaim, neuronName string) {
	// var neuronIndex, ekuiperIndex int

	// for index, container := range deploy.Spec.Template.Spec.Containers {
	// 	if container.Name == neuronName {
	// 		neuronIndex = index
	// 	}
	// }

	// if neuronIndex != 0 && ekuiperIndex != 0 {
	// 	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
	// 		Name: "shared-tmp",
	// 		VolumeSource: corev1.VolumeSource{
	// 			EmptyDir: &corev1.EmptyDirVolumeSource{},
	// 		},
	// 	})
	// 	deploy.Spec.Template.Spec.Containers[neuronIndex].VolumeMounts =
	// 		append(deploy.Spec.Template.Spec.Containers[neuronIndex].VolumeMounts, corev1.VolumeMount{
	// 			Name:      "shared-tmp",
	// 			MountPath: "/tmp",
	// 		})
	// 	deploy.Spec.Template.Spec.Containers[ekuiperIndex].VolumeMounts =
	// 		append(deploy.Spec.Template.Spec.Containers[ekuiperIndex].VolumeMounts, corev1.VolumeMount{
	// 			Name:      "shared-tmp",
	// 			MountPath: "/tmp",
	// 		})
	// }

	if len(claimList) != 0 {
		for _, pvc := range claimList {
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: pvc.Name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvc.Name,
					},
				},
			})
		}
	} else {
		deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: "neuron-data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		})
	}

	sub.updateStorageNeuron(deploy)
}

func (sub neuronDeploy) updateStorageNeuron(deploy *appsv1.Deployment) {
	neuron := deploy.Spec.Template.Spec.Containers[0]
	for _, volume := range deploy.Spec.Template.Spec.Volumes {
		if strings.Contains(volume.Name, "neuron-data") {
			neuron.VolumeMounts = append(neuron.VolumeMounts, corev1.VolumeMount{
				Name:      volume.Name,
				MountPath: "/opt/neuron/persistence",
			})
		}
	}
	deploy.Spec.Template.Spec.Containers[0] = neuron
}
