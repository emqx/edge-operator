package controllers

import (
	"context"
	"encoding/json"

	emperror "emperror.dev/errors"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ekuiperTool struct{}

func (sub ekuiperTool) reconcile(ctx context.Context, r *NeuronEXReconciler, instance *edgev1alpha1.NeuronEX) *requeue {
	cm := sub.getConfigMap(instance)
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cm), cm); err != nil {
		if k8sErrors.IsNotFound(err) {
			if err := create(ctx, r, instance, cm); err != nil {
				return &requeue{curError: emperror.Wrapf(err, "failed to create configmap")}
			}
		}
		return &requeue{curError: emperror.Wrap(err, "failed to get PVC")}
	}
	return nil
}

func (sub ekuiperTool) updateDeployment(deploy *appsv1.Deployment, instance *edgev1alpha1.NeuronEX) {
	tool := sub.getEkuiperToolContainer(instance.Spec.EKuiper)
	deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, *tool)

	cm := sub.getConfigMap(instance)
	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "ekuiper-tool-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cm.GetName(),
				},
			},
		},
	})
}

func (sub ekuiperTool) getConfigMap(instance *edgev1alpha1.NeuronEX) *corev1.ConfigMap {
	neuronStream := map[string]interface{}{
		"command": map[string]interface{}{
			"url":         "/streams",
			"description": "create neuronStream",
			"method":      "post",
			"data": map[string]string{
				"sql": `create stream neuronStream() WITH (TYPE="neuron",FORMAT="json",SHARED="true");`,
			},
		},
	}
	str, _ := json.Marshal(neuronStream)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.GetName() + "ekuiper-tool-config",
			Namespace:   instance.GetNamespace(),
			Labels:      instance.GetLabels(),
			Annotations: instance.GetAnnotations(),
		},
		Data: map[string]string{
			"neuronStream.json": string(str),
		},
	}

	cm.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))
	return cm
}

func (sub ekuiperTool) getEkuiperToolContainer(ekuiper corev1.Container) *corev1.Container {
	return &corev1.Container{
		Name:            "ekuiper-tool",
		Image:           "lfedge/ekuiper-kubernetes-tool:latest",
		ImagePullPolicy: ekuiper.ImagePullPolicy,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "ekuiper-tool-config",
				MountPath: "/kuiper-kubernetes-tool/sample",
			},
		},
	}
}
