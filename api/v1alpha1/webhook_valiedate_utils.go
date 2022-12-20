package v1alpha1

import (
	"errors"
	"reflect"
	"strings"
)

func validateNeuronContainer(ins EdgeInterface) error {
	neuron := ins.GetNeuron()

	if neuron.Image == "" {
		return errors.New("neuron container image is empty")
	}

	return nil
}

func validateEKuiperContainer(ins EdgeInterface) error {
	ekuiper := ins.GetEKuiper()

	if ekuiper.Image == "" {
		return errors.New("ekuiper container image is empty")
	}

	if !strings.HasSuffix(ekuiper.Image, "-slim-python") && !strings.HasSuffix(ekuiper.Image, "-slim") {
		return errors.New("ekuiper container image must be slim or slim-python")
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
