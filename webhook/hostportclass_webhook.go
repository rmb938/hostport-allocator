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
var hostportclasslog = logf.Log.WithName("hostportclass-resource")

func SetupHostPortClassWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.HostPortClass{}).
		WithValidator(&HostPortClassValidator{}).
		WithDefaulter(&HostPortClassDefaulter{}).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-hostport-rmb938-com-v1alpha1-hostportclass,mutating=true,failurePolicy=fail,groups=hostport.rmb938.com,resources=hostportclasses,verbs=create;update,versions=v1alpha1,sideEffects=None,admissionReviewVersions=v1,name=mhostportclass.kb.io

type HostPortClassDefaulter struct{}

var _ webhook.CustomDefaulter = &HostPortClassDefaulter{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (d *HostPortClassDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	r, ok := obj.(*v1alpha1.HostPortClass)
	if !ok {
		return fmt.Errorf("expected a HostPortClassDefaulter object but got %T", obj)
	}

	hostportclasslog.Info("default", "name", r.Name)

	if r.DeletionTimestamp.IsZero() {
		controllerutil.AddFinalizer(r, v1alpha1.HostPortFinalizer)
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-hostport-rmb938-com-v1alpha1-hostportclass,mutating=false,failurePolicy=fail,groups=hostport.rmb938.com,resources=hostportclasses,versions=v1alpha1,sideEffects=None,admissionReviewVersions=v1,name=vhostportclass.kb.io

type HostPortClassValidator struct{}

var _ webhook.CustomValidator = &HostPortClassValidator{}

func (d *HostPortClassValidator) validatePools(r *v1alpha1.HostPortClass) field.ErrorList {
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
func (d *HostPortClassValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	r, ok := obj.(*v1alpha1.HostPortClass)
	if !ok {
		return nil, fmt.Errorf("expected a HostPortClass object but got %T", obj)
	}

	hostportclasslog.Info("validate create", "name", r.Name)

	allErrs := d.validatePools(r)

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(
		schema.GroupKind{Group: v1alpha1.GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (d *HostPortClassValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) (admission.Warnings, error) {
	r, ok := new.(*v1alpha1.HostPortClass)
	if !ok {
		return nil, fmt.Errorf("expected a HostPortClass new object but got %T", new)
	}

	hostportclasslog.Info("validate update", "name", r.Name)
	_, ok = old.(*v1alpha1.HostPortClass)
	if !ok {
		return nil, fmt.Errorf("expected a .HostPortClass old object but got %T", old)
	}

	allErrs := d.validatePools(r)

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(
		schema.GroupKind{Group: v1alpha1.GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (d *HostPortClassValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	r, ok := obj.(*v1alpha1.HostPortClass)
	if !ok {
		return nil, fmt.Errorf("expected a HostPortClass object but got %T", obj)
	}

	hostportclasslog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
