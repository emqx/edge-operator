package controllers

import (
	"context"

	emperror "emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type addNeuronPVC struct{}

func (sub addNeuronPVC) reconcile(ctx context.Context, r *NeuronReconciler, volumeClaimTemplate *corev1.PersistentVolumeClaim) *requeue {
	for _, pvc := range sub.getClaimList(volumeClaimTemplate) {
		if err := r.Client.Get(ctx, client.ObjectKeyFromObject(pvc), pvc); err != nil {
			if k8sErrors.IsNotFound(err) {
				if err := r.Client.Create(ctx, pvc); err != nil {
					return &requeue{curError: emperror.Wrap(err, "failed to create PVC")}
				}
			}
			return &requeue{curError: emperror.Wrap(err, "failed to get PVC")}
		}
	}
	return nil
}

func (sub addNeuronPVC) getClaimList(volumeClaimTemplate *corev1.PersistentVolumeClaim) []*corev1.PersistentVolumeClaim {
	if volumeClaimTemplate == nil {
		return nil
	}
	return []*corev1.PersistentVolumeClaim{
		sub.addNeuronDataClaim(volumeClaimTemplate),
	}
}

func (sub addNeuronPVC) GetNeuronClaimList(volumeClaimTemplate *corev1.PersistentVolumeClaim) []*corev1.PersistentVolumeClaim {
	if volumeClaimTemplate == nil {
		return nil
	}
	return []*corev1.PersistentVolumeClaim{
		sub.addNeuronDataClaim(volumeClaimTemplate),
	}
}

func (sub addNeuronPVC) addNeuronDataClaim(volumeClaimTemplate *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
	return sub.addClaim(volumeClaimTemplate, volumeClaimTemplate.Name+"-neuron-data")
}

func (sub addNeuronPVC) addClaim(volumeClaimTemplate *corev1.PersistentVolumeClaim, name string) *corev1.PersistentVolumeClaim {
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
