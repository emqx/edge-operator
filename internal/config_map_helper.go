package internal

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	jsoniter "github.com/json-iterator/go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	EKuiperToolConfig string = "ekuiper-tool-config"
)

type ConfigMapInfo struct {
	// pod.volumeMount.name
	MountName string
	// pod.volumeMount.mountPath
	MountPath string
	// spec.volume.name the full config map name is {hdb.name}-suffix
	MapNameSuffix string
	Data          map[string]string
}

var ConfigMaps = map[string]*ConfigMapInfo{
	EKuiperToolConfig: {
		MountName:     EKuiperToolConfig,
		MountPath:     "/kuiper-kubernetes-tool/sample",
		MapNameSuffix: "ekuiper-tool-config",
		Data: map[string]string{
			"neuronStream.json": getEKuiperToolConfig(),
		},
	},
}

func getEKuiperToolConfig() string {
	config := map[string]any{
		"command": map[string]interface{}{
			"url":         "/streams",
			"description": "create neuronStream",
			"method":      "post",
			"data": map[string]string{
				"sql": "create stream neuronStream() WITH (TYPE=\"neuron\",FORMAT=\"json\",SHARED=\"true\");",
			},
		},
	}
	res, _ := jsoniter.MarshalToString(config)
	return res
}

func GetConfigMap(ins client.Object, configName string, compType edgev1alpha1.ComponentType) corev1.ConfigMap {
	cmi, has := ConfigMaps[configName]
	if !has {
		panic("no such config map name " + configName)
	}

	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: GetObjectMetadata(ins),
		Data:       cmi.Data,
	}
	cm.Name = GetResNameOnPanic(ins, cmi.MapNameSuffix)

	return cm
}
