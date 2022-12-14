package controllers

import (
	"context"

	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
)

type addEkuiperService struct{}

func (a addEkuiperService) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.EKuiper) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add eKuiper Service")

	return addService(ctx, r, instance, logger)
}

type addNeuronService struct{}

func (a addNeuronService) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.Neuron) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler", "add Neuron Service")

	return addService(ctx, r, instance, logger)
}

type addNeuronExService struct{}

func (a addNeuronExService) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.NeuronEX) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add NeuronEx Service")

	return addService(ctx, r, instance, logger)
}

func addService(ctx context.Context, r *EdgeController, ins edgev1alpha1.EdgeInterface, logger logr.Logger) *requeue {
	if ins.GetServiceTemplate() == nil {
		return nil
	}
	svc := ins.GetServiceTemplate().DeepCopy()
	svc.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))

	if err := r.createOrUpdate(ctx, ins, svc, logger); err != nil {
		return &requeue{curError: err}
	}
	return nil
}
