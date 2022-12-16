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
var ekuiperlog = logf.Log.WithName("EKuiper Webhook")

func (r *EKuiper) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-edge-emqx-io-v1alpha1-ekuiper,mutating=true,failurePolicy=fail,sideEffects=None,groups=edge.emqx.io,resources=ekuipers,verbs=create;update,versions=v1alpha1,name=mutate.ekuiper.edge.emqx.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &EKuiper{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *EKuiper) Default() {
	ekuiperlog.Info("Set default value", "name", r.Name)

	setDefaultLabels(r)
	mergeEnv(r.GetEKuiper(), &defEKuiper)
	setContainerPortsFromEnv(r.GetEKuiper())
	setDefaultService(r)
	setDefaultVolume(r)
	setDefaultEKuiperProbe(r)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-edge-emqx-io-v1alpha1-ekuiper,mutating=false,failurePolicy=fail,sideEffects=None,groups=edge.emqx.io,resources=ekuipers,verbs=create;update,versions=v1alpha1,name=validate.ekuiper.edge.emqx.io,admissionReviewVersions=v1

var _ webhook.Validator = &EKuiper{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *EKuiper) ValidateCreate() error {
	ekuiperlog.Info("validate create", "name", r.Name)

	if err := validateEKuiperContainer(r); err != nil {
		neuronexlog.Error(err, "validate ekuiper container failed")
		return err
	}

	ekuiperlog.Info("validate create success", "name", r.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *EKuiper) ValidateUpdate(old runtime.Object) error {
	ekuiperlog.Info("validate update", "name", r.Name)

	if err := validateEKuiperContainer(r); err != nil {
		neuronexlog.Error(err, "validate ekuiper container failed")
		return err
	}

	ekuiperlog.Info("validate update success", "name", r.Name)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *EKuiper) ValidateDelete() error {
	ekuiperlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
