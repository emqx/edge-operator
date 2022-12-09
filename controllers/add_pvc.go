package controllers

import (
	"context"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/emqx/edge-operator/internal"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type pvcInfo struct {
	name      string
	mountPath string
}

var defaultPVC = map[edgev1alpha1.ComponentType][]pvcInfo{
	edgev1alpha1.ComponentTypeNeuronEx: {
		{name: "neuron-data", mountPath: "/opt/neuron/persistence"},
		{name: "ekuiper-data", mountPath: "/kuiper/data"},
		{name: "ekuiper-plugin", mountPath: "/kuiper/plugins/portable"},
		{name: "shared-tmp", mountPath: "/tmp"},
	},
	edgev1alpha1.ComponentTypeEKuiper: {
		{name: "ekuiper-data", mountPath: "/kuiper/data"},
		{name: "ekuiper-plugin", mountPath: "/ekuiper-plugin"},
	},
	edgev1alpha1.ComponentTypeNeuron: {
		{name: "neuron-data", mountPath: "/opt/neuron/persistence"},
	},
}

type addEKuiperPVC struct{}

func (a addEKuiperPVC) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.EKuiper) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add eKuiper PVC")

	if !usePVC(instance) {
		return nil
	}
	return addPVC(ctx, r, instance, edgev1alpha1.ComponentTypeEKuiper, logger)
}

type addNeuronPVC struct{}

func (a addNeuronPVC) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.Neuron) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add Neuron PVC")

	if !usePVC(instance) {
		return nil
	}
	return addPVC(ctx, r, instance, edgev1alpha1.ComponentTypeNeuron, logger)
}

type addNeuronExPVC struct{}

func (a addNeuronExPVC) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.NeuronEX) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add NeuronEx PVC")

	if !usePVC(instance) {
		return nil
	}
	return addPVC(ctx, r, instance, edgev1alpha1.ComponentTypeNeuronEx, logger)
}

func addPVC(ctx context.Context, r *EdgeController, ins edgev1alpha1.EdgeInterface, compType edgev1alpha1.ComponentType,
	logger logr.Logger) *requeue {

	pvcs := defaultPVC[compType]
	for i := range pvcs {
		pvc := internal.GetPVC(ins, pvcs[i].name)
		existingPVC := &corev1.PersistentVolumeClaim{}
		err := r.Get(ctx, client.ObjectKeyFromObject(&pvc), existingPVC)
		if err != nil {
			if !k8sErrors.IsNotFound(err) {
				return &requeue{curError: err}
			}

			logger.Info("Creating PVC", "name", pvc.Name)
			// pvc no need to set ControllerReference and LastAppliedAnnotation
			if err = r.Create(ctx, &pvc); err != nil {
				if internal.IsQuotaExceeded(err) {
					return &requeue{curError: err, delayedRequeue: true}
				}
				return &requeue{curError: err}
			}
		}
	}
	return nil
}
