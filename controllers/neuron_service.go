package controllers

import (
	"context"

	emperror "emperror.dev/errors"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type neuronService struct{}

func newNeuronService() neuronService {
	return neuronService{}
}

func (sub neuronService) reconcile(ctx context.Context, r *NeuronReconciler, instance *edgev1alpha1.Neuron) *requeue {
	if instance.Spec.ServiceTemplate == nil {
		return nil
	}

	svc := &corev1.Service{
		ObjectMeta: instance.Spec.ServiceTemplate.ObjectMeta,
		Spec:       instance.Spec.ServiceTemplate.Spec,
	}
	svc.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))

	if err := createOrUpdateNeuron(ctx, r, instance, svc); err != nil {
		return &requeue{curError: emperror.Wrap(err, "failed to create or update service")}
	}
	return nil
}
