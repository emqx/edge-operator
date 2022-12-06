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

type neuronEXDeploy struct {
	pvcSubReconciler addPVC
}

func newNeuronEXDeploy() neuronEXDeploy {
	return neuronEXDeploy{
		pvcSubReconciler: addPVC{},
	}
}

func (sub neuronEXDeploy) reconcile(ctx context.Context, r *NeuronEXReconciler, instance *edgev1alpha1.NeuronEX) *requeue {
	if err := sub.pvcSubReconciler.reconcile(ctx, r, instance.Spec.VolumeClaimTemplate); err != nil {
		return err
	}

	deploy := sub.getDeployment(instance)
	sub.updateStorage(deploy, sub.pvcSubReconciler.getClaimList(instance.Spec.VolumeClaimTemplate), instance.Spec.Neuron.Name, instance.Spec.EKuiper.Name)

	if err := createOrUpdate(ctx, r, instance, deploy); err != nil {
		return &requeue{curError: emperror.Wrap(err, "failed to create or update deployment")}
	}

	return nil
}

func (sub neuronEXDeploy) getDeployment(instance *edgev1alpha1.NeuronEX) *appsv1.Deployment {
	labels := instance.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	delete(labels, "kubectl.kubernetes.io/last-applied-configuration")

	return &appsv1.Deployment{
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
}

func (sub neuronEXDeploy) getPodSpec(instance *edgev1alpha1.NeuronEX) corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{
			*sub.getNeuronContainer(instance.Spec.Neuron.DeepCopy()),
			*sub.getEkuiperContainer(instance.Spec.EKuiper.DeepCopy()),
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

func (sub neuronEXDeploy) getNeuronContainer(neuron *corev1.Container) *corev1.Container {
	return neuron
}

func (sub neuronEXDeploy) getEkuiperContainer(ekuiper *corev1.Container) *corev1.Container {
	ekuiper.Env = append([]corev1.EnvVar{
		{
			Name:  "MQTT_SOURCE__DEFAULT__SERVER",
			Value: "tcp://broker.emqx.io:1883",
		},
		{
			Name:  "KUIPER__BASIC__FILELOG",
			Value: "false",
		},
		{
			Name:  "KUIPER__BASIC__CONSOLELOG",
			Value: "true",
		},
	}, ekuiper.Env...)

	return ekuiper
}

func (sub neuronEXDeploy) updateStorage(deploy *appsv1.Deployment, claimList []*corev1.PersistentVolumeClaim, neuronName, ekuiperName string) {
	var neuronIndex, ekuiperIndex int

	for index, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == neuronName {
			neuronIndex = index
		}
		if container.Name == ekuiperName {
			ekuiperIndex = index
		}
	}

	if neuronIndex != 0 && ekuiperIndex != 0 {
		deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: "shared-tmp",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
		deploy.Spec.Template.Spec.Containers[neuronIndex].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[neuronIndex].VolumeMounts, corev1.VolumeMount{
			Name:      "shared-tmp",
			MountPath: "/tmp",
		})
		deploy.Spec.Template.Spec.Containers[ekuiperIndex].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[ekuiperIndex].VolumeMounts, corev1.VolumeMount{
			Name:      "shared-tmp",
			MountPath: "/tmp",
		})
	}

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
		if neuronIndex != 0 {
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: "neuron-data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
			})
		}
		if ekuiperIndex != 0 {
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, []corev1.Volume{
				{Name: "ekuiper-data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "ekuiper-plugin", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			}...)
		}
	}

	if neuronIndex != 0 {
		sub.updateStorageNeuron(deploy, neuronIndex)
	}
	if ekuiperIndex != 0 {
		sub.updateStorageEkuiper(deploy, ekuiperIndex)
	}

}

func (sub neuronEXDeploy) updateStorageNeuron(deploy *appsv1.Deployment, neuronIndex int) {
	neuron := deploy.Spec.Template.Spec.Containers[neuronIndex]
	for _, volume := range deploy.Spec.Template.Spec.Volumes {
		if strings.Contains(volume.Name, "neuron-data") {
			neuron.VolumeMounts = append(neuron.VolumeMounts, corev1.VolumeMount{
				Name:      volume.Name,
				MountPath: "/opt/neuron/persistence",
			})
		}
	}
	deploy.Spec.Template.Spec.Containers[neuronIndex] = neuron
}

func (sub neuronEXDeploy) updateStorageEkuiper(deploy *appsv1.Deployment, ekuiperIndex int) {
	ekuiper := deploy.Spec.Template.Spec.Containers[ekuiperIndex]
	for _, volume := range deploy.Spec.Template.Spec.Volumes {
		if strings.Contains(volume.Name, "ekuiper-data") {
			ekuiper.VolumeMounts = append(ekuiper.VolumeMounts, corev1.VolumeMount{
				Name:      volume.Name,
				MountPath: "/kuiper/data",
			})
		}

		if strings.Contains(volume.Name, "ekuiper-plugin") {
			ekuiper.VolumeMounts = append(ekuiper.VolumeMounts, corev1.VolumeMount{
				Name:      volume.Name,
				MountPath: "/kuiper/plugins/portable",
			})
		}

	}
	deploy.Spec.Template.Spec.Containers[ekuiperIndex] = ekuiper
}
