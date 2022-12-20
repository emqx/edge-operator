package internal

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetObjectMetadata returns the ObjectMetadata for a component
func GetObjectMetadata(ins metav1.Object, name string) metav1.ObjectMeta {
	metadata := &metav1.ObjectMeta{
		Name:        name,
		Namespace:   ins.GetNamespace(),
		Labels:      ins.GetLabels(),
		Annotations: ins.GetAnnotations(),
	}
	delete(metadata.Annotations, corev1.LastAppliedConfigAnnotation)
	if metadata.Namespace == "" {
		metadata.Namespace = corev1.NamespaceDefault
	}
	return *metadata
}

func GetResNameOnPanic(ins metav1.Object, shortName string) string {
	if shortName == "" {
		panic("short name is empty")
	}
	return ins.GetName() + "-" + shortName
}
