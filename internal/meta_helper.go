package internal

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetObjectMetadata returns the ObjectMetadata for a component
func GetObjectMetadata(ins client.Object, name string) metav1.ObjectMeta {
	metadata := &metav1.ObjectMeta{
		Name:        name,
		Namespace:   ins.GetNamespace(),
		Labels:      ins.GetLabels(),
		Annotations: ins.GetAnnotations(),
	}
	delete(metadata.Annotations, corev1.LastAppliedConfigAnnotation)
	return *metadata
}

func GetResNameOnPanic(ins client.Object, shortName string) string {
	if shortName == "" {
		panic("short name is empty")
	}
	return ins.GetName() + "-" + shortName
}
