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

package webhooks

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	hostportv1alpha1 "github.com/rmb938/hostport-allocator/api/v1alpha1"
	"github.com/rmb938/hostport-allocator/webhook"
	"github.com/rmb938/hostport-allocator/webhook/admission"
)

// log is for logging in this package.
var hostportclasslog = logf.Log.WithName("hostportclass-resource")

type HostPortClassWebhook struct {
	client client.Client
}

func (w *HostPortClassWebhook) SetupWebhookWithManager(mgr ctrl.Manager) {
	w.client = mgr.GetClient()
	hookServer := mgr.GetWebhookServer()

	hookServer.Register("/mutate-hostport-rmb938-com-v1alpha1-hostportclass", admission.DefaultingWebhookFor(w, &hostportv1alpha1.HostPortClass{}))
	hookServer.Register("/validate-hostport-rmb938-com-v1alpha1-hostportclass", admission.ValidatingWebhookFor(w, &hostportv1alpha1.HostPortClass{}))
}

var _ webhook.Defaulter = &HostPortClassWebhook{}

func (w *HostPortClassWebhook) Default(obj runtime.Object) error {
	r := obj.(*hostportv1alpha1.HostPortClass)

	hostportclasslog.Info("default", "name", r.Name)

	if r.DeletionTimestamp.IsZero() {
		controllerutil.AddFinalizer(r, hostportv1alpha1.HostPortFinalizer)
	}

	return nil
}

var _ webhook.Validator = &HostPortClassWebhook{}

func (w *HostPortClassWebhook) validatePools(r *hostportv1alpha1.HostPortClass) field.ErrorList {
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

func (w *HostPortClassWebhook) ValidateCreate(obj runtime.Object) error {
	_ = context.Background()
	r := obj.(*hostportv1alpha1.HostPortClass)

	hostportclasslog.Info("validate create", "name", r.Name)

	allErrs := w.validatePools(r)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: hostportv1alpha1.GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)

}

func (w *HostPortClassWebhook) ValidateUpdate(obj runtime.Object, old runtime.Object) error {
	_ = context.Background()
	r := obj.(*hostportv1alpha1.HostPortClass)

	hostportclasslog.Info("validate update", "name", r.Name)
	_ = old.(*hostportv1alpha1.HostPortClass)

	allErrs := w.validatePools(r)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: hostportv1alpha1.GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)

}

func (w *HostPortClassWebhook) ValidateDelete(obj runtime.Object) error {
	r := obj.(*hostportv1alpha1.HostPortClass)

	hostportclasslog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
