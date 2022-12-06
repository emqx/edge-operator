package controllers

import (
	"context"

	emperror "emperror.dev/errors"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type neuronEXService struct{}

func newNeuronEXService() neuronEXService {
	return neuronEXService{}
}

func (sub neuronEXService) reconcile(ctx context.Context, r *NeuronEXReconciler, instance *edgev1alpha1.NeuronEX) *requeue {
	if instance.Spec.ServiceTemplate == nil {
		return nil
	}

	svc := &corev1.Service{
		ObjectMeta: instance.Spec.ServiceTemplate.ObjectMeta,
		Spec:       instance.Spec.ServiceTemplate.Spec,
	}
	svc.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))

	if err := createOrUpdate(ctx, r, instance, svc); err != nil {
		return &requeue{curError: emperror.Wrap(err, "failed to create or update service")}
	}
	return nil
}
