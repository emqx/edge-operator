package controllers

import (
	"context"
	"github.com/emqx/edge-operator/internal"
	jsoniter "github.com/json-iterator/go"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type addEkuiperTool struct{}

func (a addEkuiperTool) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.NeuronEX) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add eKuiper tool")

	cm, has := internal.ConfigMaps.Get(internal.EKuiperToolConfig)
	if !has {
		panic("no such config map " + internal.EKuiperToolConfig)
	}

	defConfig := internal.GetEKuiperToolConfig()
	file, _ := json.MarshalToString(defConfig)

	newConfigMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: internal.GetObjectMetadata(instance, nil, instance.GetComponentType()),
		Data: map[string]string{
			// neuronStream.json: file
			cm.MapKey: file,
		},
	}
	newConfigMap.Name = internal.GetResNameOnPanic(instance, cm.MapNameSuffix)

	existingConfigMap := &corev1.ConfigMap{}
	if err := r.Get(ctx, client.ObjectKeyFromObject(newConfigMap), existingConfigMap); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return &requeue{curError: err}
		}

		logger.Info("Creating ConfigMap", "name", newConfigMap.Name)
		if err = r.create(ctx, instance, newConfigMap); err != nil {
			return &requeue{curError: err}
		}
	}
	return nil
}
