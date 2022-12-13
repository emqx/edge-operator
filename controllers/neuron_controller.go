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
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NeuronReconciler reconciles a NeuronEX object
type NeuronReconciler struct {
	*EdgeController
}

func NewNeuronReconciler(mgr manager.Manager) *NeuronReconciler {
	return &NeuronReconciler{
		EdgeController: NewEdgeController(mgr),
	}
}

//+kubebuilder:rbac:groups=edge.emqx.io,resources=neurons,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=edge.emqx.io,resources=neurons/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=edge.emqx.io,resources=neurons/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the NeuronEX object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NeuronReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconcile(ctx, req, &edgev1alpha1.Neuron{})
}

// SetupWithManager sets up the controller with the Manager.
func (r *NeuronReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&edgev1alpha1.Neuron{}).
		Complete(r)
}