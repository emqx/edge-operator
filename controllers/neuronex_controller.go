/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

type Patcher struct {
	*patch.Annotator
	patch.Maker
}

type subReconciler interface {
	reconcile(ctx context.Context, r *NeuronEXReconciler, instance edgev1alpha1.EdgeInterface) *requeue
}

// NeuronEXReconciler reconciles a NeuronEX object
type NeuronEXReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	Recorder          record.EventRecorder
	Patcher           *Patcher
	subReconcilerList []subReconciler
}

func NewNeuronEXReconciler(mgr manager.Manager) *NeuronEXReconciler {
	var patcher *Patcher = new(Patcher)
	patcher.Annotator = patch.NewAnnotator(edgev1alpha1.GroupVersion.Group + "/last-applied-configuration")
	patcher.Maker = patch.NewPatchMaker(
		patcher.Annotator,
		&patch.K8sStrategicMergePatcher{},
		&patch.BaseJSONMergePatcher{},
	)

	return &NeuronEXReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("neuronEX-controller"),
		Patcher:  patcher,
		subReconcilerList: []subReconciler{
			newNeuronEXDeploy(),
			newNeuronEXService(),
		},
	}
}

//+kubebuilder:rbac:groups=edge.emqx.io,resources=neuronexs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=edge.emqx.io,resources=neuronexs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=edge.emqx.io,resources=neuronexs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NeuronEX object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NeuronEXReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("neuronEX", req.NamespacedName)

	instance := &edgev1alpha1.NeuronEX{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if k8sErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if instance.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	delayedRequeue := false
	for _, subReconciler := range r.subReconcilerList {
		requeue := subReconciler.reconcile(ctx, r, instance)
		if requeue == nil {
			continue
		}

		if requeue.delayedRequeue {
			logger.Info("Delaying requeue for sub-reconciler",
				"subReconciler", fmt.Sprintf("%T", subReconciler),
				"message", requeue.message,
				"error", requeue.curError)
			delayedRequeue = true
			continue
		}
		return processRequeue(requeue, subReconciler, instance, r.Recorder, logger)
	}

	if delayedRequeue {
		logger.Info("NeuronEX was not fully reconciled by reconciliation process")
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NeuronEXReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&edgev1alpha1.NeuronEX{}).
		Complete(r)
}
