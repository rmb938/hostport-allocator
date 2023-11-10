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
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
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
var hostportlog = logf.Log.WithName("hostport-resource")

func (r *HostPort) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-hostport-rmb938-com-v1alpha1-hostport,mutating=true,failurePolicy=fail,groups=hostport.rmb938.com,resources=hostports,verbs=create;update,versions=v1alpha1,sideEffects=None,admissionReviewVersions=v1,name=mhostport.kb.io

var _ webhook.Defaulter = &HostPort{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *HostPort) Default() {
	hostportlog.Info("default", "name", r.Name)

	if r.DeletionTimestamp.IsZero() {
		controllerutil.AddFinalizer(r, HostPortFinalizer)
	}

	return
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-hostport-rmb938-com-v1alpha1-hostport,mutating=false,failurePolicy=fail,groups=hostport.rmb938.com,resources=hostports,versions=v1alpha1,sideEffects=None,admissionReviewVersions=v1,name=vhostport.kb.io

var _ webhook.Validator = &HostPort{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *HostPort) ValidateCreate() (admission.Warnings, error) {
	hostportlog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(
		schema.GroupKind{Group: GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *HostPort) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {

	hostportlog.Info("validate update", "name", r.Name)
	oldHP := old.(*HostPort)

	var allErrs field.ErrorList

	// don't allow changing class
	if r.Spec.HostPortClassName != oldHP.Spec.HostPortClassName {
		allErrs = append(allErrs,
			field.Forbidden(field.NewPath("spec").Child("hostPortClassName"),
				"cannot change hostPortClassName"),
		)
	}

	// don't allow changing claim
	if equality.Semantic.DeepEqual(oldHP.Spec.ClaimRef, r.Spec.ClaimRef) == false {
		allErrs = append(allErrs,
			field.Forbidden(field.NewPath("spec").Child("claimRef"),
				"cannot change claimRef"),
		)
	}

	// don't allow changing port once set
	if oldHP.Status.Port > 0 && r.Status.Port != oldHP.Status.Port {
		allErrs = append(allErrs,
			field.Forbidden(field.NewPath("status").Child("port"),
				"cannot change port"),
		)
	}

	// TODO: only allow setting port when also setting as allocated
	if oldHP.Status.Port == 0 && r.Status.Port > 0 && r.Status.Phase != HostPortPhaseAllocated {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("status").Child("port"), r.Status.Port,
				fmt.Sprintf("port can only be set when also setting the phase to %s", HostPortPhaseAllocated)),
		)
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(
		schema.GroupKind{Group: GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *HostPort) ValidateDelete() (admission.Warnings, error) {
	hostportlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
