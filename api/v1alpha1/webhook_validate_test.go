package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
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
					Name:  "neuron",
					Image: "emqx/neuron:latest",
				},
				EKuiper: corev1.Container{
					Name:  "ekuiper",
					Image: "lfedge/ekuiper:latest-slim",
				},
				ServiceTemplate:     &corev1.Service{},
				VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{},
			},
		},
		&Neuron{
			ObjectMeta: metav1.ObjectMeta{
				Name: "neuron",
			},
			Spec: NeuronSpec{
				Neuron: corev1.Container{
					Name:  "neuron",
					Image: "emqx/neuron:latest",
				},
				ServiceTemplate:     &corev1.Service{},
				VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{},
			},
		},
		&EKuiper{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ekuiper",
			},
			Spec: EKuiperSpec{
				EKuiper: corev1.Container{
					Name:  "ekuiper",
					Image: "lfedge/ekuiper:latest-slim",
				},
				ServiceTemplate:     &corev1.Service{},
				VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{},
			},
		},
	} {
		t.Run("check success", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			assert.Nil(t, got.ValidateCreate())
			assert.Nil(t, got.ValidateUpdate(ins))
			assert.Nil(t, got.ValidateDelete())
		})
		t.Run("check container name is empty", func(t *testing.T) {
			if ins.GetNeuron() != nil {
				got := deepCopyEdgeEdgeInterface(ins)
				got.GetNeuron().Name = ""
				assert.ErrorContains(t, got.ValidateCreate(), "neuron container name is empty")
				assert.ErrorContains(t, got.ValidateUpdate(ins), "neuron container name is empty")
			}
			if ins.GetEKuiper() != nil {
				got := deepCopyEdgeEdgeInterface(ins)
				got.GetEKuiper().Name = ""
				assert.ErrorContains(t, got.ValidateCreate(), "ekuiper container name is empty")
				assert.ErrorContains(t, got.ValidateUpdate(ins), "ekuiper container name is empty")
			}
		})
		t.Run("check container image is empty", func(t *testing.T) {
			if ins.GetNeuron() != nil {
				got := deepCopyEdgeEdgeInterface(ins)
				got.GetNeuron().Image = ""
				assert.ErrorContains(t, got.ValidateCreate(), "neuron container image is empty")
				assert.ErrorContains(t, got.ValidateUpdate(ins), "neuron container image is empty")
			}
			if ins.GetEKuiper() != nil {
				got := deepCopyEdgeEdgeInterface(ins)
				got.GetEKuiper().Image = ""
				assert.ErrorContains(t, got.ValidateCreate(), "ekuiper container image is empty")
				assert.ErrorContains(t, got.ValidateUpdate(ins), "ekuiper container image is empty")
			}
		})
		t.Run("check ekuiper image is not slim or slim-python", func(t *testing.T) {
			if ins.GetEKuiper() != nil {
				got := deepCopyEdgeEdgeInterface(ins)
				got.GetEKuiper().Image = "lfedge/ekuiper:latest"
				assert.ErrorContains(t, got.ValidateCreate(), "ekuiper container image must be slim or slim-python")
				assert.ErrorContains(t, got.ValidateUpdate(ins), "ekuiper container image must be slim or slim-python")
			}
		})
	}
}
