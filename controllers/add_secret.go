package controllers

import (
	"context"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/emqx/edge-operator/internal"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
)

type addEKuiperSecret struct{}

func (a addEKuiperSecret) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.EKuiper) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add eKuiper Secret")
	return addSecret(ctx, r, instance, logger)
}

type addNeuronSecret struct{}

func (a addNeuronSecret) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.Neuron) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add Neuron Secret")
	return addSecret(ctx, r, instance, logger)
}

type addNeuronExSecret struct{}

func (a addNeuronExSecret) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.NeuronEX) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add NeuronEx Secret")
	return addSecret(ctx, r, instance, logger)
}

func addSecret(ctx context.Context, r *EdgeController, ins edgev1alpha1.EdgeInterface, logger logr.Logger) *requeue {
	secret := corev1.Secret{
		Type:       corev1.SecretTypeOpaque,
		ObjectMeta: internal.GetObjectMetadata(ins, internal.GetResNameOnPanic(ins, publicKey)),
		Data:       map[string][]byte{},
	}
	secret.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))

	publicKeys := ins.GetEdgePodSpec().PublicKeys
	for i := range publicKeys {
		pk := &publicKeys[i]
		secret.Data[pk.Name] = []byte(pk.Data)
	}
	if err := r.createOrUpdate(ctx, ins, &secret, logger); err != nil {
		return &requeue{curError: err}
	}
	return nil
}
