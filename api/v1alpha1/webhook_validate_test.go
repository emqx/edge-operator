package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValiedate(t *testing.T) {
	for _, ins := range []EdgeInterface{
		&NeuronEX{
			ObjectMeta: metav1.ObjectMeta{
				Name: "neuronex",
			},
			Spec: NeuronEXSpec{
				Neuron: corev1.Container{
					Image: "emqx/neuron:latest",
				},
				EKuiper: corev1.Container{
					Image: "lfedge/ekuiper:latest-slim",
				},
			},
		},
		&Neuron{
			ObjectMeta: metav1.ObjectMeta{
				Name: "neuron",
			},
			Spec: NeuronSpec{
				Neuron: corev1.Container{
					Image: "emqx/neuron:latest",
				},
			},
		},
		&EKuiper{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ekuiper",
			},
			Spec: EKuiperSpec{
				EKuiper: corev1.Container{
					Image: "lfedge/ekuiper:latest-slim",
				},
			},
		},
	} {
		t.Run("check success", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			assert.Nil(t, got.ValidateCreate())
			assert.Nil(t, got.ValidateUpdate(ins))
			assert.Nil(t, got.ValidateDelete())
		})
		t.Run("check volume template is empty", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			got.SetVolumeClaimTemplate(nil)
			assert.Nil(t, got.ValidateCreate())

			got.SetVolumeClaimTemplate(&corev1.PersistentVolumeClaimTemplate{})
			assert.ErrorContains(t, got.ValidateCreate(), "volume template access modes is empty")

			got.SetVolumeClaimTemplate(&corev1.PersistentVolumeClaimTemplate{
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
				},
			})
			assert.ErrorContains(t, got.ValidateCreate(), "volume template resources storage is empty")

			got.SetVolumeClaimTemplate(&corev1.PersistentVolumeClaimTemplate{
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("8Mi"),
						},
					},
				},
			})
			assert.Nil(t, got.ValidateCreate())
		})
		t.Run("check volume template can not be updated", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			assert.Nil(t, got.ValidateUpdate(ins))

			got.SetVolumeClaimTemplate(&corev1.PersistentVolumeClaimTemplate{
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("8Mi"),
						},
					},
				},
			})
			assert.ErrorContains(t, got.ValidateUpdate(ins), "volume template can not be updated")
		})
	}
}
