package controllers

import (
	"context"

	emperror "emperror.dev/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createOrUpdate(ctx context.Context, r *NeuronEXReconciler, owner, obj client.Object) error {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
	if err := r.Client.Get(context.TODO(), client.ObjectKeyFromObject(obj), u); err != nil {
		if k8sErrors.IsNotFound(err) {
			return create(ctx, r, owner, obj)
		}
		return emperror.Wrapf(err, "failed to get %s %s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
	}

	patcherResult, err := r.Patcher.Calculate(u, obj)
	if err != nil {
		return emperror.Wrapf(err, "failed to calculate patch for %s %s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
	}
	if !patcherResult.IsEmpty() {
		return update(ctx, r, owner, obj, u)
	}

	return nil
}

func create(ctx context.Context, r *NeuronEXReconciler, owner, obj client.Object) error {
	if err := ctrl.SetControllerReference(owner, obj, r.Scheme); err != nil {
		return emperror.Wrapf(err, "failed to set controller reference for %s %s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
	}
	if err := r.Patcher.SetLastAppliedAnnotation(obj); err != nil {
		return emperror.Wrapf(err, "failed to set last applied annotation for %s %s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
	}
	if err := r.Client.Create(ctx, obj); err != nil {
		return emperror.Wrapf(err, "failed to create %s %s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
	}
	return nil
}

func update(ctx context.Context, r *NeuronEXReconciler, owner, curObj, oriObj client.Object) error {
	annotations := curObj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	for key, value := range oriObj.GetAnnotations() {
		if _, present := annotations[key]; !present {
			annotations[key] = value
		}
	}
	curObj.SetAnnotations(annotations)
	curObj.SetResourceVersion(oriObj.GetResourceVersion())
	curObj.SetCreationTimestamp(oriObj.GetCreationTimestamp())
	curObj.SetManagedFields(oriObj.GetManagedFields())

	if err := ctrl.SetControllerReference(owner, curObj, r.Scheme); err != nil {
		return emperror.Wrapf(err, "failed to set controller reference for %s %s", curObj.GetObjectKind().GroupVersionKind().Kind, curObj.GetName())
	}
	if err := r.Patcher.SetLastAppliedAnnotation(curObj); err != nil {
		return emperror.Wrapf(err, "failed to set controller reference for %s %s", curObj.GetObjectKind().GroupVersionKind().Kind, curObj.GetName())
	}
	if err := r.Client.Update(context.TODO(), curObj); err != nil {
		return emperror.Wrapf(err, "failed to update %s %s", curObj.GetObjectKind().GroupVersionKind().Kind, curObj.GetName())
	}

	return nil
}
