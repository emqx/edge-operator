package internal

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type metaObject interface {
	metav1.ObjectMetaAccessor
	client.Object
}

func GetDeployment(ins metaObject, compType edgev1alpha1.ComponentType, podSpec *corev1.PodTemplateSpec) appsv1.Deployment {
	meta := ins.GetObjectMeta().(*metav1.ObjectMeta)

	deploy := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: GetObjectMetadata(ins, meta, compType),
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: podSpec.GetLabels(),
			},
			Template: *podSpec,
		},
	}

	deploy.Name = compType.GetResName(ins)
	return deploy
}
