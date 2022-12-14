package internal

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetService builds a service.
func GetService(ins client.Object, svc *corev1.Service) corev1.Service {
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: GetObjectMetadata(ins, &svc.ObjectMeta),
		Spec:       *svc.Spec.DeepCopy(),
	}

	service.Name = svc.Name
	return service
}
