package internal

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetService builds a service.
func GetService(ins client.Object, svc *corev1.Service, compType edgev1alpha1.ComponentType) corev1.Service {
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: GetObjectMetadata(ins, &svc.ObjectMeta, compType),
		Spec:       *svc.Spec.DeepCopy(),
	}

	if len(service.Spec.Selector) == 0 {
		service.Spec.Selector = make(map[string]string)
	}
	return service
}
