package controllers

import (
	"context"
	"fmt"

	emperror "emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var log = logf.Log.WithName("Edge Controller")

type CR interface {
	*edgev1alpha1.EKuiper | *edgev1alpha1.NeuronEX | *edgev1alpha1.Neuron
}

type subReconciler[T CR] interface {
	reconcile(ctx context.Context, r *EdgeController, instance T) *requeue
}

type patcher struct {
	patch.Maker
	*patch.Annotator
}

type EdgeController struct {
	client.Client
	patcher  *patcher
	Recorder record.EventRecorder
}

func NewEdgeController(mgr manager.Manager) *EdgeController {
	annotator := patch.NewAnnotator(edgev1alpha1.GroupVersion.Group + "/last-applied-configuration")

	return &EdgeController{
		Client:   mgr.GetClient(),
		Recorder: mgr.GetEventRecorderFor("neuronEX-controller"),
		patcher: &patcher{
			Maker: patch.NewPatchMaker(
				annotator,
				&patch.K8sStrategicMergePatcher{},
				&patch.BaseJSONMergePatcher{},
			),
			Annotator: annotator,
		},
	}
}

func (ec *EdgeController) reconcile(ctx context.Context, req ctrl.Request, cr client.Object) (ctrl.Result, error) {
	if err := ec.Get(ctx, req.NamespacedName, cr); err != nil {
		if k8sErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if cr.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	switch cr.(type) {
	case *edgev1alpha1.NeuronEX:
		subs := []subReconciler[*edgev1alpha1.NeuronEX]{
			updateNeuronEXStatus{},
			addEkuiperTool{},
			addNeuronExPVC{},
			addNeuronExDeploy{},
			addNeuronExService{},
			updateNeuronEXStatus{},
		}
		return subReconcile[*edgev1alpha1.NeuronEX](ec, ctx, cr, subs)
	case *edgev1alpha1.EKuiper:
		subs := []subReconciler[*edgev1alpha1.EKuiper]{
			updateEkuiperStatus{},
			addEKuiperPVC{},
			addEkuiperDeployment{},
			addEkuiperService{},
			updateEkuiperStatus{},
		}
		return subReconcile[*edgev1alpha1.EKuiper](ec, ctx, cr, subs)
	case *edgev1alpha1.Neuron:
		subs := []subReconciler[*edgev1alpha1.Neuron]{
			updateNeuronStatus{},
			addNeuronPVC{},
			addNeuronDeployment{},
			addNeuronService{},
			updateNeuronStatus{},
		}
		return subReconcile[*edgev1alpha1.Neuron](ec, ctx, cr, subs)
	default:
		panic("unknown kind")
	}
}

func subReconcile[T CR](ec *EdgeController, ctx context.Context, obj client.Object, subReconcilers []subReconciler[T]) (
	ctrl.Result, error) {

	logger := log.WithValues("namespace", obj.GetNamespace(), "instance", obj.GetName())
	logger.Info("reconcile")

	delayedRequeue := false
	for _, subReconciler := range subReconcilers {
		logger.Info("Attempting to run sub-reconciler", "subReconciler", fmt.Sprintf("%T", subReconciler))
		requeue := subReconciler.reconcile(ctx, ec, obj.(T))
		if requeue == nil {
			continue
		}

		if requeue.delayedRequeue {
			logger.Info("Delaying requeue for sub-reconciler",
				"kind", obj.GetObjectKind().GroupVersionKind().String(),
				"subReconciler", fmt.Sprintf("%T", subReconciler),
				"message", requeue.message,
				"error", requeue.curError)
			delayedRequeue = true
			continue
		}
		return processRequeue(requeue, subReconciler, obj, ec.Recorder, logger)
	}

	if delayedRequeue {
		logger.Info("not fully reconciled by reconciliation process", "kind", obj.GetObjectKind())
		return ctrl.Result{Requeue: true}, nil
	}

	logger.Info("Reconciliation complete", "kind", obj.GetObjectKind().GroupVersionKind().String())
	ec.Recorder.Event(obj, corev1.EventTypeNormal, "ReconciliationComplete", "")

	return ctrl.Result{}, nil
}

func (ec *EdgeController) createOrUpdate(ctx context.Context, owner, newObj client.Object, logger logr.Logger) error {
	existingObj := &unstructured.Unstructured{}
	existingObj.SetGroupVersionKind(newObj.GetObjectKind().GroupVersionKind())

	if err := ec.Get(ctx, client.ObjectKeyFromObject(newObj), existingObj); err != nil {
		if k8sErrors.IsNotFound(err) {
			logger.Info("Create", "res", newObj.GetName())
			return ec.create(ctx, owner, newObj)
		}
		return emperror.Wrapf(err, "failed to get %s %s", newObj.GetObjectKind().GroupVersionKind().Kind, newObj.GetName())
	}

	patcherResult, err := ec.patcher.Calculate(existingObj, newObj)
	if err != nil {
		return emperror.Wrapf(err, "failed to calculate patch for %s %s", newObj.GetObjectKind().GroupVersionKind().Kind, newObj.GetName())
	}
	if !patcherResult.IsEmpty() {
		logger.Info("Update", "res", newObj.GetName())
		return ec.update(ctx, owner, newObj, existingObj)
	}
	return nil
}

func (ec *EdgeController) create(ctx context.Context, owner, newObj client.Object) error {
	if err := ctrl.SetControllerReference(owner, newObj, ec.Scheme()); err != nil {
		return emperror.Wrapf(err, "failed to set controller reference for %s %s", newObj.GetObjectKind().GroupVersionKind().Kind, newObj.GetName())
	}
	if err := ec.patcher.SetLastAppliedAnnotation(newObj); err != nil {
		return emperror.Wrapf(err, "failed to set last applied annotation for %s %s", newObj.GetObjectKind().GroupVersionKind().Kind, newObj.GetName())
	}
	if err := ec.Create(ctx, newObj); err != nil {
		return emperror.Wrapf(err, "failed to create %s %s", newObj.GetObjectKind().GroupVersionKind().Kind, newObj.GetName())
	}
	return nil
}

func (ec *EdgeController) update(ctx context.Context, owner, newObj, existingObj client.Object) error {
	annotations := existingObj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	for key, value := range newObj.GetAnnotations() {
		annotations[key] = value
	}

	newObj.SetAnnotations(annotations)
	newObj.SetResourceVersion(existingObj.GetResourceVersion())
	newObj.SetCreationTimestamp(existingObj.GetCreationTimestamp())
	newObj.SetManagedFields(existingObj.GetManagedFields())

	if err := ctrl.SetControllerReference(owner, newObj, ec.Scheme()); err != nil {
		return emperror.Wrapf(err, "failed to set controller reference for %s %s",
			newObj.GetObjectKind().GroupVersionKind().Kind, newObj.GetName())
	}

	if err := ec.patcher.SetLastAppliedAnnotation(newObj); err != nil {
		return emperror.Wrapf(err, "failed to set last applied annotation for %s %s",
			newObj.GetObjectKind().GroupVersionKind().Kind, newObj.GetName())
	}

	if err := ec.Update(ctx, newObj); err != nil {
		return emperror.Wrapf(err, "failed to update %s %s",
			newObj.GetObjectKind().GroupVersionKind().Kind, newObj.GetName())
	}

	return nil
}
