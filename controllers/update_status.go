package controllers

import (
	"context"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type updateEkuiperStatus struct{}

func (u updateEkuiperStatus) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.EKuiper) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"update status")

	return updateStatus(ctx, r, instance, logger)
}

type updateNeuronStatus struct{}

func (u updateNeuronStatus) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.Neuron) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"update status")

	return updateStatus(ctx, r, instance, logger)
}

type updateNeuronEXStatus struct{}

func (u updateNeuronEXStatus) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.NeuronEX) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"update status")

	return updateStatus(ctx, r, instance, logger)
}

func updateStatus(ctx context.Context, r *EdgeController, instance edgev1alpha1.EdgeInterface, logger logr.Logger) *requeue {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: instance.GetNamespace(),
			Name:      instance.GetName(),
		},
	}

	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(deploy), deploy); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return &requeue{curError: err}
		}
		return nil
	}

	phase := edgev1alpha1.CRNotReady
	if deploy.Status.ReadyReplicas == deploy.Status.Replicas {
		phase = edgev1alpha1.CRReady
	}

	if instance.GetStatus().Phase != phase {
		instance.SetStatus(
			&edgev1alpha1.EdgeStatus{
				Phase: phase,
			},
		)
		logger.Info("Update status", "current", instance.GetStatus())
		if err := r.Status().Update(ctx, instance); err != nil {
			return &requeue{curError: err}
		}
	}
	return nil
}
