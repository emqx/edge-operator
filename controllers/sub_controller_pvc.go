package controllers

import (
	"context"

	emperror "emperror.dev/errors"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type subPVC struct{}

func (sub subPVC) reconcile(ctx context.Context, r edgeReconcilerInterface, instance edgev1alpha1.EdgeInterface) *requeue {
	for _, pvc := range sub.getClaimMap(instance) {
		if err := r.Get(ctx, client.ObjectKeyFromObject(pvc), pvc); err != nil {
			if k8sErrors.IsNotFound(err) {
				if err := r.Create(ctx, pvc); err != nil {
					return &requeue{curError: emperror.Wrap(err, "failed to create PVC")}
				}
			}
			return &requeue{curError: emperror.Wrap(err, "failed to get PVC")}
		}
	}
	return nil
}

func (sub subPVC) updateDeployment(deploy *appsv1.Deployment, instance edgev1alpha1.EdgeInterface) {
	var neuronIndex, ekuiperIndex *int

	for index, container := range deploy.Spec.Template.Spec.Containers {
		if instance.GetNeuron() != nil {
			if container.Name == instance.GetNeuron().Name {
				*neuronIndex = index
			}
		}
		if instance.GetEKuiper() != nil {
			if container.Name == instance.GetEKuiper().Name {
				*ekuiperIndex = index
			}
		}
	}

	claimMap := sub.getClaimMap(instance)

	if neuronIndex != nil {
		deploy.Spec.Template.Spec.Containers[*neuronIndex].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[*neuronIndex].VolumeMounts, corev1.VolumeMount{
			Name:      "neuron-data",
			MountPath: "/opt/neuron/persistence",
		})
		if pvc, ok := claimMap["neuron-data"]; ok {
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: "neuron-data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvc.Name,
					},
				},
			})
		} else {
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: "neuron-data",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
		}
	}

	if ekuiperIndex != nil {
		deploy.Spec.Template.Spec.Containers[*ekuiperIndex].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[*ekuiperIndex].VolumeMounts, corev1.VolumeMount{
			Name:      "ekuiper-data",
			MountPath: "/kuiper/data",
		})

		if pvc, ok := claimMap["ekuiper-data"]; ok {
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: "ekuiper-data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvc.Name,
					},
				},
			})
		} else {
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: "ekuiper-data",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
		}

		deploy.Spec.Template.Spec.Containers[*ekuiperIndex].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[*ekuiperIndex].VolumeMounts, corev1.VolumeMount{
			Name:      "ekuiper-plugin",
			MountPath: "/kuiper/plugins/portable",
		})

		if pvc, ok := claimMap["ekuiper-plugin"]; ok {
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: "ekuiper-plugin",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvc.Name,
					},
				},
			})
		} else {
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: "ekuiper-plugin",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
		}
	}

	if neuronIndex != nil && ekuiperIndex != nil {
		deploy.Spec.Template.Spec.Containers[*neuronIndex].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[*neuronIndex].VolumeMounts, corev1.VolumeMount{
			Name:      "shared-tmp",
			MountPath: "/tmp",
		})
		deploy.Spec.Template.Spec.Containers[*ekuiperIndex].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[*ekuiperIndex].VolumeMounts, corev1.VolumeMount{
			Name:      "shared-tmp",
			MountPath: "/tmp",
		})
		deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: "shared-tmp",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}
}

func (sub subPVC) getClaimMap(instance edgev1alpha1.EdgeInterface) map[string]*corev1.PersistentVolumeClaim {
	if instance.GetVolumeClaimTemplate() != nil {
		return nil
	}

	claimMap := make(map[string]*corev1.PersistentVolumeClaim)

	if instance.GetNeuron() != nil {
		claimMap["neuron-data"] = sub.addNeuronDataClaim(instance.GetVolumeClaimTemplate())
	}

	if instance.GetEKuiper() != nil {
		claimMap["ekuiper-data"] = sub.addEkuiperDataClaim(instance.GetVolumeClaimTemplate())
		claimMap["ekuiper-plugin"] = sub.addEkuiperPluginClaim(instance.GetVolumeClaimTemplate())
	}

	return claimMap
}

func (sub subPVC) addNeuronDataClaim(volumeClaimTemplate *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
	return sub.addClaim(volumeClaimTemplate, volumeClaimTemplate.Name+"-neuron-data")
}

func (sub subPVC) addEkuiperDataClaim(volumeClaimTemplate *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
	return sub.addClaim(volumeClaimTemplate, volumeClaimTemplate.Name+"-ekuiper-data")
}

func (sub subPVC) addEkuiperPluginClaim(volumeClaimTemplate *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
	return sub.addClaim(volumeClaimTemplate, volumeClaimTemplate.Name+"-ekuiper-plugin")
}

func (sub subPVC) addClaim(volumeClaimTemplate *corev1.PersistentVolumeClaim, name string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   volumeClaimTemplate.Namespace,
			Labels:      volumeClaimTemplate.ObjectMeta.Labels,
			Annotations: volumeClaimTemplate.Annotations,
		},
		Spec: volumeClaimTemplate.Spec,
	}
}
