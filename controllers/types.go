package controllers

import (
	"context"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type edgeReconcilerInterface interface {
	client.Client
	patcherInterface
}

type patcherInterface interface {
	patch.Maker
	SetLastAppliedAnnotation(runtime.Object) error
}

type patcher struct {
	patch.Maker
	*patch.Annotator
}

type subReconciler interface {
	reconcile(ctx context.Context, r edgeReconcilerInterface, instance edgev1alpha1.EdgeInterface) *requeue
}
