/*


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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var hostportclasslog = logf.Log.WithName("hostportclass-resource")

func (r *HostPortClass) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-hostport-rmb938-com-v1alpha1-hostportclass,mutating=true,failurePolicy=fail,groups=hostport.rmb938.com,resources=hostportclasses,verbs=create;update,versions=v1alpha1,sideEffects=None,admissionReviewVersions=v1,name=mhostportclass.kb.io

var _ webhook.Defaulter = &HostPortClass{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *HostPortClass) Default() {
	hostportclasslog.Info("default", "name", r.Name)

	if r.DeletionTimestamp.IsZero() {
		controllerutil.AddFinalizer(r, HostPortFinalizer)
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-hostport-rmb938-com-v1alpha1-hostportclass,mutating=false,failurePolicy=fail,groups=hostport.rmb938.com,resources=hostportclasses,versions=v1alpha1,sideEffects=None,admissionReviewVersions=v1,name=vhostportclass.kb.io

var _ webhook.Validator = &HostPortClass{}

func (r *HostPortClass) validatePools() field.ErrorList {
	var allErrs field.ErrorList

	for index, pool := range r.Spec.Pools {
		if pool.Start > pool.End {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("pools").Index(index).Child("end"), pool.End,
				"End must be greater or equal to start"))
		}
	}

	// TODO: make sure there are no overlapping pools

	return allErrs
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *HostPortClass) ValidateCreate() (admission.Warnings, error) {
	hostportclasslog.Info("validate create", "name", r.Name)

	allErrs := r.validatePools()

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(
		schema.GroupKind{Group: GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *HostPortClass) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	hostportclasslog.Info("validate update", "name", r.Name)
	_ = old.(*HostPortClass)

	allErrs := r.validatePools()

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(
		schema.GroupKind{Group: GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *HostPortClass) ValidateDelete() (admission.Warnings, error) {
	hostportclasslog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
