package v1alpha1

import (
	"errors"
	"reflect"
)

func validateNeuronContainer(ins EdgeInterface) error {
	neuron := ins.GetNeuron()

	if neuron.Image == "" {
		return errors.New("neuron container image is empty")
	}

	return nil
}

func validateVolumeTemplateCreate(ins EdgeInterface) error {
	vol := ins.GetVolumeClaimTemplate()
	if vol == nil {
		return nil
	}

	if len(vol.Spec.AccessModes) == 0 {
		return errors.New("volume template access modes is empty")
	}
	if vol.Spec.Resources.Limits.Storage().IsZero() && vol.Spec.Resources.Requests.Storage().IsZero() {
		return errors.New("volume template resources storage is empty")
	}

	return nil
}

func validateVolumeTemplateUpdate(new, old EdgeInterface) error {
	if !reflect.DeepEqual(new.GetVolumeClaimTemplate(), old.GetVolumeClaimTemplate()) {
		return errors.New("volume template can not be updated")
	}
	return nil
}
