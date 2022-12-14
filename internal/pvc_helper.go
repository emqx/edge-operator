package internal

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPVC(ins edgev1alpha1.EdgeInterface, shortName string) (pvc corev1.PersistentVolumeClaim) {
	if ins.GetVolumeClaimTemplate() != nil {
		pvc = *ins.GetVolumeClaimTemplate().DeepCopy()
	}

	pvc.ObjectMeta = getPvcMetadata(ins)
	pvc.ObjectMeta.Name = GetPvcName(ins, shortName)

	if pvc.Spec.AccessModes == nil {
		pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	}

	if pvc.Spec.Resources.Requests == nil {
		pvc.Spec.Resources.Requests = corev1.ResourceList{}
	}

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
					Name: GetResNameOnPanic(ins, m.MapNameSuffix)},
			},
		},
	}
}

// getPvcMetadata returns the metadata for a PVC
func getPvcMetadata(ins edgev1alpha1.EdgeInterface) metav1.ObjectMeta {
	var customMetadata *metav1.ObjectMeta

	if ins.GetVolumeClaimTemplate() != nil {
		customMetadata = &ins.GetVolumeClaimTemplate().ObjectMeta
	}
	return GetObjectMetadata(ins, customMetadata)
}
