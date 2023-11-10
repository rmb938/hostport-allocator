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
var hostportclaimlog = logf.Log.WithName("hostportclaim-resource")

func (r *HostPortClaim) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-hostport-rmb938-com-v1alpha1-hostportclaim,mutating=true,failurePolicy=fail,groups=hostport.rmb938.com,resources=hostportclaims,verbs=create;update,versions=v1alpha1,sideEffects=None,admissionReviewVersions=v1,name=mhostportclaim.kb.io

var _ webhook.Defaulter = &HostPortClaim{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *HostPortClaim) Default() {

	hostportclaimlog.Info("default", "name", r.Name)

	if r.DeletionTimestamp.IsZero() {
		controllerutil.AddFinalizer(r, HostPortFinalizer)
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-hostport-rmb938-com-v1alpha1-hostportclaim,mutating=false,failurePolicy=fail,groups=hostport.rmb938.com,resources=hostportclaims,versions=v1alpha1,sideEffects=None,admissionReviewVersions=v1,name=vhostportclaim.kb.io

var _ webhook.Validator = &HostPortClaim{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *HostPortClaim) ValidateCreate() (admission.Warnings, error) {

	hostportclaimlog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(
		schema.GroupKind{Group: GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *HostPortClaim) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	hostportclaimlog.Info("validate update", "name", r.Name)
	oldHPC := old.(*HostPortClaim)

	var allErrs field.ErrorList

	if r.Spec.HostPortClassName != oldHPC.Spec.HostPortClassName {
		allErrs = append(allErrs,
			field.Forbidden(field.NewPath("spec").Child("hostPortClassName"),
				"cannot change hostPortClassName"),
		)
	}

	if len(oldHPC.Spec.HostPortName) > 0 && oldHPC.Spec.HostPortName != r.Spec.HostPortName {
		allErrs = append(allErrs,
			field.Forbidden(field.NewPath("spec").Child("hostPortName"),
				"cannot change hostPortName once set"),
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
func (r *HostPortClaim) ValidateDelete() (admission.Warnings, error) {
	hostportclaimlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
