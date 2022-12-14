package internal

import (
	"strings"

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

	if metadata.Labels == nil {
		metadata.Labels = make(map[string]string)
	}

	if metadata.Annotations == nil {
		metadata.Annotations = make(map[string]string)
	}
	return *metadata
}

func GetResNameOnPanic(ins client.Object, shortName string) string {
	if shortName == "" {
		panic("short name is empty")
	}
	return GetResNameWithDefault(ins, shortName, "")
}

// GetResNameWithDefault get resource name with short name, it will use default name if short name is empty
func GetResNameWithDefault(ins client.Object, shortName, defaultName string) string {
	buf := strings.Builder{}
	buf.WriteString(ins.GetName())
	buf.WriteRune('-')
	if shortName == "" {
		buf.WriteString(defaultName)
	} else {
		buf.WriteString(shortName)
	}
	return buf.String()
}
