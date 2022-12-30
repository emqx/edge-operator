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
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// EKuiperReconciler reconciles a EKuiper object
type EKuiperReconciler struct {
	*EdgeController
}

func NewEKuiperReconciler(k8sClient client.Client, eventRecorder record.EventRecorder) *EKuiperReconciler {
	return &EKuiperReconciler{
		EdgeController: NewEdgeController(k8sClient, eventRecorder),
	}
}

//+kubebuilder:rbac:groups=edge.emqx.io,resources=ekuipers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=edge.emqx.io,resources=ekuipers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=edge.emqx.io,resources=ekuipers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the EKuiper object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile

func (r *EKuiperReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconcile(ctx, req, &edgev1alpha1.EKuiper{})
}

// SetupWithManager sets up the controller with the Manager.
func (r *EKuiperReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&edgev1alpha1.EKuiper{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
