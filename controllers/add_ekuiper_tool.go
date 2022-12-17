package controllers

import (
	"context"

	"github.com/emqx/edge-operator/internal"
	jsoniter "github.com/json-iterator/go"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type addEkuiperTool struct{}

func (a addEkuiperTool) reconcile(ctx context.Context, r *EdgeController, ins *edgev1alpha1.NeuronEX) *requeue {
	logger := log.WithValues("namespace", ins.Namespace, "instance", ins.Name, "reconciler",
		"add eKuiper tool")

	volume := getEkuiperToolCOnfigVol(ins)

	cnfigMap := &corev1.ConfigMap{
		ObjectMeta: internal.GetObjectMetadata(ins, volume.volumeSource.ConfigMap.Name),
		Data: map[string]string{
			"neuronStream.json": getekuiperToolConfig(),
		},
	}

	existingConfigMap := &corev1.ConfigMap{}
	if err := r.Get(ctx, client.ObjectKeyFromObject(cnfigMap), existingConfigMap); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return &requeue{curError: err}
		}

		logger.Info("Creating ConfigMap", "name", cnfigMap.Name)
		if err = r.create(ctx, ins, cnfigMap); err != nil {
			return &requeue{curError: err}
		}
	}
	return nil
}

func getekuiperToolConfig() string {
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
