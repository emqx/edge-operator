package controllers

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/emqx/edge-operator/internal"
	corev1 "k8s.io/api/core/v1"
)

type mountTo string

const (
	mountToNeuron      mountTo = "neuron"
	mountToEkuiper     mountTo = "ekuiper"
	mountToEkuiperTool mountTo = "ekuiper-tool"
)

const (
	neuronData        string = "neuron-data"
	ekuiperData       string = "ekuiper-data"
	ekuiperPlugins    string = "ekuiper-plugins"
	ekuiperToolConfig string = "ekuiper-tool-config"
	sharedTmp         string = "shared-tmp"
)

type volumeInfo struct {
	name         string
	mountPath    string
	mountTo      []mountTo
	volumeSource corev1.VolumeSource
}

func getNeuronDataVol(ins edgev1alpha1.EdgeInterface) volumeInfo {
	v := volumeInfo{
		name:      neuronData,
		mountPath: "/opt/neuron/persistence",
		mountTo:   []mountTo{mountToNeuron},
	}
	if ins.GetVolumeClaimTemplate() != nil {
		v.volumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: internal.GetResNameOnPanic(ins.GetVolumeClaimTemplate(), v.name),
			},
		}
		return v
	}
	v.volumeSource = corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{},
	}
	return v
}

func getEKuiperDataVol(ins edgev1alpha1.EdgeInterface) volumeInfo {
	v := volumeInfo{
		name:      ekuiperData,
		mountPath: "/kuiper/data",
		mountTo:   []mountTo{mountToEkuiper},
	}
	if ins.GetVolumeClaimTemplate() != nil {
		v.volumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: internal.GetResNameOnPanic(ins.GetVolumeClaimTemplate(), v.name),
			},
		}
		return v
	}
	v.volumeSource = corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{},
	}
	return v
}

func getEKuiperPluginsVol(ins edgev1alpha1.EdgeInterface) volumeInfo {
	v := volumeInfo{
		name:      ekuiperPlugins,
		mountPath: "/kuiper/plugins/portable",
		mountTo:   []mountTo{mountToEkuiper},
	}
	if ins.GetVolumeClaimTemplate() != nil {
		v.volumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: internal.GetResNameOnPanic(ins.GetVolumeClaimTemplate(), v.name),
			},
		}
		return v
	}
	v.volumeSource = corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{},
	}
	return v
}

func getEkuiperToolCOnfigVol(ins edgev1alpha1.EdgeInterface) volumeInfo {
	return volumeInfo{
		name:      ekuiperToolConfig,
		mountPath: "/kuiper-kubernetes-tool/sample",
		mountTo:   []mountTo{mountToEkuiperTool},
		volumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: internal.GetResNameOnPanic(ins, ekuiperToolConfig),
				},
				DefaultMode: &[]int32{corev1.ConfigMapVolumeSourceDefaultMode}[0],
			},
		},
	}
}

func getShardTmpVol(ins edgev1alpha1.EdgeInterface) volumeInfo {
	return volumeInfo{
		name:      sharedTmp,
		mountPath: "/tmp",
		mountTo:   []mountTo{mountToNeuron, mountToEkuiper},
		volumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

func getVolumeList(ins edgev1alpha1.EdgeInterface) []volumeInfo {
	list := []volumeInfo{}
	switch ins.(type) {
	case *edgev1alpha1.NeuronEX:
		list = append(list, getNeuronDataVol(ins), getEKuiperDataVol(ins), getEKuiperPluginsVol(ins), getEkuiperToolCOnfigVol(ins), getShardTmpVol(ins))
	case *edgev1alpha1.Neuron:
		list = append(list, getNeuronDataVol(ins))
	case *edgev1alpha1.EKuiper:
		list = append(list, getEKuiperDataVol(ins), getEKuiperPluginsVol(ins))
	}
	return list
}
