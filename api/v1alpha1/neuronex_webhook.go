/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var neuronexlog = logf.Log.WithName("NeuronEX Webhook")

func (r *NeuronEX) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-edge-emqx-io-v1alpha1-neuronex,mutating=true,failurePolicy=fail,sideEffects=None,groups=edge.emqx.io,resources=neuronexs,verbs=create;update,versions=v1alpha1,name=mutate.neuronex.edge.emqx.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &NeuronEX{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *NeuronEX) Default() {
	neuronexlog.Info("Set default value", "name", r.Name)

	setDefaultLabels(r)

	mergeEnv(r.GetNeuron(), &defNeuron)
	mergeContainerPorts(r.GetNeuron(), &defNeuron)

	mergeEnv(r.GetEKuiper(), &defEKuiper)
	setContainerPortsFromEnv(r.GetEKuiper())

	setDefaultService(r)
	setDefaultVolume(r)

	setDefaultNeuronProbe(r)
	setDefaultEKuiperProbe(r)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-edge-emqx-io-v1alpha1-neuronex,mutating=false,failurePolicy=fail,sideEffects=None,groups=edge.emqx.io,resources=neuronexs,verbs=create;update,versions=v1alpha1,name=validate.neuronex.edge.emqx.io,admissionReviewVersions=v1

var _ webhook.Validator = &NeuronEX{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *NeuronEX) ValidateCreate() error {
	neuronexlog.Info("validate create", "name", r.Name)

	if err := validateNeuronContainer(r); err != nil {
		neuronexlog.Error(err, "validate neuron container failed")
		return err
	}

	if err := validateEKuiperContainer(r); err != nil {
		neuronexlog.Error(err, "validate ekuiper container failed")
		return err
	}

	neuronexlog.Info("validate create success", "name", r.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *NeuronEX) ValidateUpdate(old runtime.Object) error {
	neuronexlog.Info("validate update", "name", r.Name)

	if err := validateNeuronContainer(r); err != nil {
		neuronexlog.Error(err, "validate neuron container failed")
		return err
	}

	if err := validateEKuiperContainer(r); err != nil {
		neuronexlog.Error(err, "validate ekuiper container failed")
		return err
	}

	neuronexlog.Info("validate update success", "name", r.Name)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *NeuronEX) ValidateDelete() error {
	neuronexlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
