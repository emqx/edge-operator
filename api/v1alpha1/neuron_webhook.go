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
var neuronlog = logf.Log.WithName("neuron-resource")

func (r *Neuron) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-edge-emqx-io-v1alpha1-neuron,mutating=true,failurePolicy=fail,sideEffects=None,groups=edge.emqx.io,resources=neurons,verbs=create;update,versions=v1alpha1,name=mutate.neuron.edge.emqx.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Neuron{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Neuron) Default() {
	neuronlog.Info("Set default value", "name", r.Name)

	defValue := Neuron{
		ObjectMeta: getCRObjectMeta(r.Name, ComponentTypeNeuron),
		Spec: NeuronSpec{
			Neuron: defNeuron,
		},
	}

	mergeLabelsInMetadata(&r.ObjectMeta, defValue.ObjectMeta)
	mergeAnnotations(&r.ObjectMeta, defValue.ObjectMeta)
	extendEnv(&r.Spec.Neuron, defValue.Spec.Neuron.Env)
	setDefaultService(r)
	setDefaultVolume(r)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-edge-emqx-io-v1alpha1-neuron,mutating=false,failurePolicy=fail,sideEffects=None,groups=edge.emqx.io,resources=neurons,verbs=create;update,versions=v1alpha1,name=vneuron.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Neuron{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Neuron) ValidateCreate() error {
	neuronlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Neuron) ValidateUpdate(old runtime.Object) error {
	neuronlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Neuron) ValidateDelete() error {
	neuronlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
