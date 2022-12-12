package internal

import (
	"strings"

	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetObjectMetadata returns the ObjectMetadata for a component
func GetObjectMetadata(ins client.Object, base *metav1.ObjectMeta, compType edgev1alpha1.ComponentType) metav1.ObjectMeta {
	metadata := &metav1.ObjectMeta{}
	if base != nil {
		metadata.Annotations = base.Annotations
		metadata.Labels = base.Labels
		delete(metadata.Labels, corev1.LastAppliedConfigAnnotation)
	}
	metadata.Namespace = ins.GetNamespace()

	if metadata.Labels == nil {
		metadata.Labels = make(map[string]string)
	}

	// TODO set default label by webhook
	metadata.Labels[edgev1alpha1.ComponentKey] = compType.String()

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
