package controllers

import (
	"context"
	"github.com/emqx/edge-operator/internal"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type addRuleSet struct{}

func (a addRuleSet) reconcile(ctx context.Context, r *EdgeController, ins *edgev1alpha1.NeuronEX) *requeue {
	logger := log.WithValues("namespace", ins.Namespace, "instance", ins.Name, "reconciler",
		"add eKuiper rule set")

	ruleSet := &corev1.ConfigMap{
		ObjectMeta: internal.GetObjectMetadata(ins, internal.GetResNameOnPanic(ins, ekuiperRuleSet)),
		Data: map[string]string{
			"init.json": `{"streams": {"neuronStream": "CREATE STREAM neuronStream () WITH (DATASOURCE=\"users\", FORMAT=\"JSON\")"}}`,
		},
	}

	existingRuleSet := &corev1.ConfigMap{}
	if err := r.Get(ctx, client.ObjectKeyFromObject(ruleSet), existingRuleSet); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return &requeue{curError: err}
		}

		logger.Info("Creating eKuiper rule set", "name", ruleSet.Name)
		if err = r.create(ctx, ins, ruleSet); err != nil {
			return &requeue{curError: err}
		}
	}
	return nil
}
