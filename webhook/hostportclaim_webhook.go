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

package webhook

import (
	"context"
	"fmt"

	"github.com/rmb938/hostport-allocator/api/v1alpha1"
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

func SetupHostPortClaimWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.HostPortClaim{}).
		WithValidator(&HostPortClaimValidator{}).
		WithDefaulter(&HostPortClaimDefaulter{}).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-hostport-rmb938-com-v1alpha1-hostportclaim,mutating=true,failurePolicy=fail,groups=hostport.rmb938.com,resources=hostportclaims,verbs=create;update,versions=v1alpha1,sideEffects=None,admissionReviewVersions=v1,name=mhostportclaim.kb.io

type HostPortClaimDefaulter struct{}

var _ webhook.CustomDefaulter = &HostPortClaimDefaulter{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (d *HostPortClaimDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	r, ok := obj.(*v1alpha1.HostPortClaim)
	if !ok {
		return fmt.Errorf("expected a HostPortClaim object but got %T", obj)
	}

	hostportclaimlog.Info("default", "name", r.Name)

	if r.DeletionTimestamp.IsZero() {
		controllerutil.AddFinalizer(r, v1alpha1.HostPortFinalizer)
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-hostport-rmb938-com-v1alpha1-hostportclaim,mutating=false,failurePolicy=fail,groups=hostport.rmb938.com,resources=hostportclaims,versions=v1alpha1,sideEffects=None,admissionReviewVersions=v1,name=vhostportclaim.kb.io

type HostPortClaimValidator struct{}

var _ webhook.CustomValidator = &HostPortClaimValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (d *HostPortClaimValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	r, ok := obj.(*v1alpha1.HostPortClaim)
	if !ok {
		return nil, fmt.Errorf("expected a HostPortClaim object but got %T", obj)
	}

	hostportclaimlog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(
		schema.GroupKind{Group: v1alpha1.GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (d *HostPortClaimValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) (admission.Warnings, error) {
	r, ok := new.(*v1alpha1.HostPortClaim)
	if !ok {
		return nil, fmt.Errorf("expected a HostPortClaim new object but got %T", new)
	}

	hostportclaimlog.Info("validate update", "name", r.Name)
	oldHPC, ok := old.(*v1alpha1.HostPortClaim)
	if !ok {
		return nil, fmt.Errorf("expected a HostPortClaim old object but got %T", old)
	}

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
		schema.GroupKind{Group: v1alpha1.GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (d *HostPortClaimValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	r, ok := obj.(*v1alpha1.HostPortClaim)
	if !ok {
		return nil, fmt.Errorf("expected a HostPortClaim object but got %T", obj)
	}

	hostportclaimlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
