package controllers

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/emqx/edge-operator/internal"
	corev1 "k8s.io/api/core/v1"
)

type mountTo = string

const (
	mountToNeuron  mountTo = "neuron"
	mountToEkuiper mountTo = "ekuiper"
)

const (
	neuronData     = "neuron-data"
	ekuiperData    = "ekuiper-data"
	ekuiperPlugins = "ekuiper-plugins"
	ekuiperRuleSet = "ekuiper-init-rule-set"
	sharedTmp      = "shared-tmp"
	publicKey      = "public-key"
)

type mountAttr struct {
	path     string
	readOnly bool
}

type volumeInfo struct {
	name         string
	mounts       map[mountTo]mountAttr
	volumeSource corev1.VolumeSource
}

func getPersistentVolumeSource(ins edgev1alpha1.EdgeInterface, name string) (volumeSource corev1.VolumeSource) {
	if ins.GetVolumeClaimTemplate() != nil {
		volumeSource.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
			ClaimName: internal.GetResNameOnPanic(ins.GetVolumeClaimTemplate(), name),
		}
		return
	}

	volumeSource.EmptyDir = &corev1.EmptyDirVolumeSource{}
	return
}

func getSecretVol(ins edgev1alpha1.EdgeInterface) volumeInfo {
	secretVol := volumeInfo{
		name: publicKey,
		mounts: map[mountTo]mountAttr{
			mountToNeuron: {
				path:     "/opt/neuron/certs",
				readOnly: true,
			},
			mountToEkuiper: {
				path:     "/kuiper/etc/mgmt",
				readOnly: true,
			},
		},
		volumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{
					{
						Secret: &corev1.SecretProjection{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: internal.GetResNameOnPanic(ins, publicKey),
							},
						},
					},
				},
				DefaultMode: &[]int32{0444}[0],
			},
		},
	}
	return secretVol
}

func getNeuronDataVol(ins edgev1alpha1.EdgeInterface) volumeInfo {
	v := volumeInfo{
		name: neuronData,
		mounts: map[mountTo]mountAttr{
			mountToNeuron: {
				path: "/opt/neuron/persistence",
			},
		},
		volumeSource: getPersistentVolumeSource(ins, neuronData),
	}
	return v
}

func getEKuiperDataVol(ins edgev1alpha1.EdgeInterface) volumeInfo {
	v := volumeInfo{
		name: ekuiperData,
		mounts: map[mountTo]mountAttr{
			mountToEkuiper: {
				path: "/kuiper/data",
			},
		},
		volumeSource: getPersistentVolumeSource(ins, ekuiperData),
	}
	return v
}

func getEKuiperPluginsVol(ins edgev1alpha1.EdgeInterface) volumeInfo {
	v := volumeInfo{
		name: ekuiperPlugins,
		mounts: map[mountTo]mountAttr{
			mountToEkuiper: {
				path: "/kuiper/plugins/portable",
			},
		},
		volumeSource: getPersistentVolumeSource(ins, ekuiperPlugins),
	}
	return v
}

func getEkuiperInitRuleSetVol(ins edgev1alpha1.EdgeInterface) volumeInfo {
	return volumeInfo{
		name: ekuiperRuleSet,
		mounts: map[mountTo]mountAttr{
			mountToEkuiper: {
				path:     "/kuiper/data/init.json",
				readOnly: true,
			},
		},
		volumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: internal.GetResNameOnPanic(ins, ekuiperRuleSet),
				},
				DefaultMode: &[]int32{0444}[0],
			},
		},
	}
}

func getShardTmpVol() volumeInfo {
	return volumeInfo{
		name: sharedTmp,
		mounts: map[mountTo]mountAttr{
			mountToNeuron: {
				path: "/tmp",
			},
			mountToEkuiper: {
				path: "/tmp",
			},
		},
		volumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

func getVolumeList(ins edgev1alpha1.EdgeInterface) []volumeInfo {
	switch ins.GetComponentType() {
	case edgev1alpha1.ComponentTypeNeuronEx:
		return []volumeInfo{
			getNeuronDataVol(ins),
			getEKuiperDataVol(ins),
			getEKuiperPluginsVol(ins),
			getEkuiperInitRuleSetVol(ins),
			getShardTmpVol(),
			getSecretVol(ins),
		}
	case edgev1alpha1.ComponentTypeNeuron:
		return []volumeInfo{
			getNeuronDataVol(ins),
			getSecretVol(ins)}
	case edgev1alpha1.ComponentTypeEKuiper:
		return []volumeInfo{
			getEKuiperDataVol(ins),
			getEKuiperPluginsVol(ins),
			getSecretVol(ins)}
	default:
		panic("Unknown component " + ins.GetComponentType())
	}
}
