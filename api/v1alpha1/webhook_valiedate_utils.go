package v1alpha1

import (
	"errors"
	"strings"
)

func validateNeuronContainer(ins EdgeInterface) error {
	neuron := ins.GetNeuron()

	if neuron.Name == "" {
		return errors.New("neuron container name is empty")
	}

	if neuron.Image == "" {
		return errors.New("neuron container image is empty")
	}

	return nil
}

func validateEKuiperContainer(ins EdgeInterface) error {
	ekuiper := ins.GetEKuiper()

	if ekuiper.Name == "" {
		return errors.New("ekuiper container name is empty")
	}

	if ekuiper.Image == "" {
		return errors.New("ekuiper container image is empty")
	}

	if !strings.HasSuffix(ekuiper.Image, "-slim-python") && !strings.HasSuffix(ekuiper.Image, "-slim") {
		return errors.New("ekuiper container image must be slim or slim-python")
	}
	return nil
}
