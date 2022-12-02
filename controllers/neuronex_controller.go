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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

// NeuronEXReconciler reconciles a NeuronEX object
type NeuronEXReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=edge.emqx.io,resources=neuronices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=edge.emqx.io,resources=neuronices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=edge.emqx.io,resources=neuronices/finalizers,verbs=update

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
	_ = log.FromContext(ctx)

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

	podList := &corev1.PodList{}
	if err := r.Client.List(
		ctx,
		podList,
		client.InNamespace(req.Namespace),
		client.MatchingLabels(instance.GetLabels()),
	); err != nil {
		return ctrl.Result{}, err
	}
	if len(podList.Items) == 0 {
		pod := addPod(instance)
		if err := ctrl.SetControllerReference(instance, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Client.Create(ctx, pod); err != nil {
			return ctrl.Result{}, err
		}
	}

	for _, pod := range podList.Items {
		controllerRef := metav1.GetControllerOf(&pod)
		if controllerRef == nil {
			continue
		}
		if controllerRef.UID == instance.UID {
			if pod.GetDeletionTimestamp() != nil {
				pod := addPod(instance)
				if err := ctrl.SetControllerReference(instance, pod, r.Scheme); err != nil {
					return ctrl.Result{}, err
				}
				if err := r.Client.Create(ctx, pod); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NeuronEXReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&edgev1alpha1.NeuronEX{}).
		Watches(
			&source.Kind{Type: &corev1.Pod{}},
			&handler.EnqueueRequestForOwner{OwnerType: &edgev1alpha1.NeuronEX{}, IsController: true},
			builder.WithPredicates(
				predicate.Funcs{
					CreateFunc:  func(e event.CreateEvent) bool { return false },
					UpdateFunc:  func(e event.UpdateEvent) bool { return e.ObjectNew.GetDeletionTimestamp() != nil },
					DeleteFunc:  func(e event.DeleteEvent) bool { return false },
					GenericFunc: func(e event.GenericEvent) bool { return false },
				},
			),
		).
		Complete(r)
}

func addPod(instance *edgev1alpha1.NeuronEX) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: instance.Name + "-",
			Namespace:    instance.Namespace,
			Labels:       instance.GetLabels(),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				instance.Spec.Neuron,
				instance.Spec.EKuiper,
			},
		},
	}

	return pod
}
