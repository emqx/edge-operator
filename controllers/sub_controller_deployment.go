package controllers

import (
	"context"
	"reflect"

	emperror "emperror.dev/errors"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deploySubReconciler interface {
	reconcile(ctx context.Context, r edgeReconcilerInterface, instance edgev1alpha1.EdgeInterface) *requeue
	updateDeployment(deploy *appsv1.Deployment, instance edgev1alpha1.EdgeInterface)
}

type subDeploy struct {
	subReconcilerList []deploySubReconciler
}

func newSubDeploy() subDeploy {
	return subDeploy{
		subReconcilerList: []deploySubReconciler{
			subPVC{},
			subEkuiperTool{},
		},
	}
}

func (sub subDeploy) reconcile(ctx context.Context, r edgeReconcilerInterface, instance edgev1alpha1.EdgeInterface) *requeue {
	deploy := sub.getDeployment(instance)

	for _, subReconciler := range sub.subReconcilerList {
		if err := subReconciler.reconcile(ctx, r, instance); err != nil {
			return err
		}
		subReconciler.updateDeployment(deploy, instance)
	}

	if err := createOrUpdate(ctx, r, instance, deploy); err != nil {
		return &requeue{curError: emperror.Wrap(err, "failed to create or update deployment")}
	}

	return nil
}

func (sub subDeploy) getDeployment(instance edgev1alpha1.EdgeInterface) *appsv1.Deployment {
	labels := instance.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	delete(labels, "kubectl.kubernetes.io/last-applied-configuration")

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.GetName(),
			Namespace:   instance.GetNamespace(),
			Annotations: instance.GetAnnotations(),
			Labels:      labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: instance.GetAnnotations(),
				},
				Spec: sub.getPodSpec(instance),
			},
		},
	}
	deploy.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("Deployment"))
	return deploy
}

func (sub subDeploy) getPodSpec(instance edgev1alpha1.EdgeInterface) corev1.PodSpec {
	podSpec := &corev1.PodSpec{}
	edgePodSpec := instance.GetEdgePodSpec()
	structAssign(podSpec, &edgePodSpec)
	if instance.GetNeuron() != nil {
		podSpec.Containers = append(podSpec.Containers, *sub.getNeuronContainer(instance.GetNeuron().DeepCopy()))
	}
	if instance.GetEKuiper() != nil {
		podSpec.Containers = append(podSpec.Containers, *sub.getEkuiperContainer(instance.GetEKuiper().DeepCopy()))
	}

	return *podSpec
}

func (sub subDeploy) getNeuronContainer(neuron *corev1.Container) *corev1.Container {
	return neuron
}

func (sub subDeploy) getEkuiperContainer(ekuiper *corev1.Container) *corev1.Container {
	ekuiper.Env = append([]corev1.EnvVar{
		{
			Name:  "MQTT_SOURCE__DEFAULT__SERVER",
			Value: "tcp://broker.emqx.io:1883",
		},
		{
			Name:  "KUIPER__BASIC__FILELOG",
			Value: "false",
		},
		{
			Name:  "KUIPER__BASIC__CONSOLELOG",
			Value: "true",
		},
	}, ekuiper.Env...)

	return ekuiper
}

// structAssign copy the value of struct from src to dist
func structAssign(dist, src interface{}) {
	dVal := reflect.ValueOf(dist).Elem()
	sVal := reflect.ValueOf(src).Elem()
	sType := sVal.Type()
	for i := 0; i < sVal.NumField(); i++ {
		// we need to check if the dist struct has the same field
		name := sType.Field(i).Name
		if ok := dVal.FieldByName(name).IsValid(); ok {
			dVal.FieldByName(name).Set(reflect.ValueOf(sVal.Field(i).Interface()))
		}
	}
}
