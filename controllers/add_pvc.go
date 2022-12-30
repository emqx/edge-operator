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

type addEKuiperPVC struct{}

func (a addEKuiperPVC) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.EKuiper) *requeue {
	if instance.GetVolumeClaimTemplate() == nil {
		return nil
	}

	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add eKuiper PVC")
	return addPVC(ctx, r, instance, logger)
}

type addNeuronPVC struct{}

func (a addNeuronPVC) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.Neuron) *requeue {
	if instance.GetVolumeClaimTemplate() == nil {
		return nil
	}

	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add Neuron PVC")
	return addPVC(ctx, r, instance, logger)
}

type addNeuronExPVC struct{}

func (a addNeuronExPVC) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.NeuronEX) *requeue {
	if instance.GetVolumeClaimTemplate() == nil {
		return nil
	}

	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add NeuronEx PVC")
	return addPVC(ctx, r, instance, logger)
}

func addPVC(ctx context.Context, r *EdgeController, ins edgev1alpha1.EdgeInterface, logger logr.Logger) *requeue {
	vols := getVolumeList(ins)
	for i := range vols {
		if vols[i].volumeSource.PersistentVolumeClaim == nil {
			continue
		}
		template := ins.GetVolumeClaimTemplate()
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: internal.GetObjectMetadata(template, internal.GetResNameOnPanic(template, vols[i].name)),
			Spec:       template.Spec,
		}

		existingPVC := &corev1.PersistentVolumeClaim{}
		err := r.Get(ctx, client.ObjectKeyFromObject(pvc), existingPVC)
		if err != nil {
			if !k8sErrors.IsNotFound(err) {
				return &requeue{curError: err}
			}

			logger.Info("Creating PVC", "name", pvc.Name)
			// pvc no need to set ControllerReference and LastAppliedAnnotation
			if err = r.Create(ctx, pvc); err != nil {
				if internal.IsQuotaExceeded(err) {
					return &requeue{curError: err, delayedRequeue: true}
				}
				return &requeue{curError: err}
			}
		}
	}
	return nil
}
