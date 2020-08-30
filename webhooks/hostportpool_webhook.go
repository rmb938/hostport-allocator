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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	hostportv1alpha1 "github.com/rmb938/hostport-allocator/api/v1alpha1"
	"github.com/rmb938/hostport-allocator/webhook"
	"github.com/rmb938/hostport-allocator/webhook/admission"
)

// log is for logging in this package.
var hostportpoollog = logf.Log.WithName("hostportpool-resource")

type HostPortPoolWebhook struct {
	client client.Client
}

func (w *HostPortPoolWebhook) SetupWebhookWithManager(mgr ctrl.Manager) {
	w.client = mgr.GetClient()
	hookServer := mgr.GetWebhookServer()

	hookServer.Register("/mutate-hostport-rmb938-com-v1alpha1-hostportpool", admission.DefaultingWebhookFor(w, &hostportv1alpha1.HostPortPool{}))
	hookServer.Register("/validate-hostport-rmb938-com-v1alpha1-hostportpool", admission.ValidatingWebhookFor(w, &hostportv1alpha1.HostPortPool{}))
}

var _ webhook.Defaulter = &HostPortPoolWebhook{}

func (w *HostPortPoolWebhook) Default(obj runtime.Object) {
	r := obj.(*hostportv1alpha1.HostPortPool)

	hostportpoollog.Info("default", "name", r.Name)

	if r.DeletionTimestamp.IsZero() {
		hasFinalizer := false

		for _, finalizer := range r.Finalizers {
			if finalizer == hostportv1alpha1.HostPortPoolFinalizer {
				hasFinalizer = true
				break
			}
		}

		if hasFinalizer == false {
			r.Finalizers = append(r.Finalizers, hostportv1alpha1.HostPortPoolFinalizer)
		}
	}
}

var _ webhook.Validator = &HostPortPoolWebhook{}

func (w *HostPortPoolWebhook) ValidateCreate(obj runtime.Object) error {
	r := obj.(*hostportv1alpha1.HostPortPool)

	hostportpoollog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	if r.Spec.Start >= r.Spec.End {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("end"), r.Spec.End,
			"End must be larger than start"))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: hostportv1alpha1.GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)
}

func (w *HostPortPoolWebhook) ValidateUpdate(obj runtime.Object, old runtime.Object) error {
	oldHPP := old.(*hostportv1alpha1.HostPortPool)
	r := obj.(*hostportv1alpha1.HostPortPool)

	hostportpoollog.Info("validate update", "name", r.Name)

	var allErrs field.ErrorList

	if oldHPP.Spec.HostPortClassName != r.Spec.HostPortClassName {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("hostPortClassName"),
			"Cannot change hostPortClassName"))
	}

	if oldHPP.Spec.Start != r.Spec.Start {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("start"),
			"Cannot change start"))
	}

	if oldHPP.Spec.End != r.Spec.End {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("end"),
			"Cannot change end"))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: hostportv1alpha1.GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)
}

func (w *HostPortPoolWebhook) ValidateDelete(obj runtime.Object) error {
	r := obj.(*hostportv1alpha1.HostPortPool)

	hostportpoollog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
