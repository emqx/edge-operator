package controllers

import (
	"context"

	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type addEkuiperStatus struct{}

func (a addEkuiperStatus) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.EKuiper) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add eKuiper Status")

	return addStatus(ctx, r, instance, logger)
}

type addNeuronStatus struct{}

func (a addNeuronStatus) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.Neuron) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add eKuiper Status")

	return addStatus(ctx, r, instance, logger)
}

type addNeuronEXStatus struct{}

func (a addNeuronEXStatus) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.NeuronEX) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add eKuiper Status")

	return addStatus(ctx, r, instance, logger)
}

func addStatus(ctx context.Context, r *EdgeController, instance edgev1alpha1.EdgeInterface, logger logr.Logger) *requeue {
	podList := &corev1.PodList{}
	if err := r.List(ctx, podList,
		client.InNamespace(instance.GetNamespace()),
		client.MatchingLabels(instance.GetLabels()),
	); err != nil {
		return nil
	}
	if len(podList.Items) == 0 {
		return nil
	}

	instance.SetStatus(
		edgev1alpha1.EdgeStatus{
			Phase: podList.Items[0].Status.Phase,
		},
	)
	logger.Info("update eKuiper status")
	if err := r.Status().Update(ctx, instance); err != nil {
		return &requeue{curError: err}
	}
	return nil
}