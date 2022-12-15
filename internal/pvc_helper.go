package internal

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func GetPVC(ins edgev1alpha1.EdgeInterface, shortName string) (pvc corev1.PersistentVolumeClaim) {
	if ins.GetVolumeClaimTemplate() != nil {
		pvc = *ins.GetVolumeClaimTemplate().DeepCopy()
	}

	pvc.ObjectMeta = GetObjectMetadata(ins.GetVolumeClaimTemplate(), GetPvcName(ins, shortName))

	storage := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if (&storage).IsZero() {
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse("128Gi")
	}
	return
}

func GetPvcName(ins edgev1alpha1.EdgeInterface, shortName string) string {
	claim := ins.GetVolumeClaimTemplate()
	if claim != nil && claim.Name != "" {
		shortName = claim.Name + "-" + shortName
	}
	return GetResNameOnPanic(ins, shortName)
}

func GetVolume(ins edgev1alpha1.EdgeInterface, m *ConfigMapInfo) corev1.Volume {
	return corev1.Volume{
		Name: m.MountName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: GetResNameOnPanic(ins, m.MapNameSuffix),
				},
			},
		},
	}
}
