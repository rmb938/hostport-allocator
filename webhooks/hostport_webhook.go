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
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
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
var hostportlog = logf.Log.WithName("hostport-resource")

type HostPortWebhook struct {
	client client.Client
}

func (w *HostPortWebhook) SetupWebhookWithManager(mgr ctrl.Manager) {
	w.client = mgr.GetClient()
	hookServer := mgr.GetWebhookServer()

	hookServer.Register("/mutate-hostport-rmb938-com-v1alpha1-hostport", admission.DefaultingWebhookFor(w, &hostportv1alpha1.HostPort{}))
	hookServer.Register("/validate-hostport-rmb938-com-v1alpha1-hostport", admission.ValidatingWebhookFor(w, &hostportv1alpha1.HostPort{}))
}

var _ webhook.Defaulter = &HostPortWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *HostPortWebhook) Default(obj runtime.Object) error {
	r := obj.(*hostportv1alpha1.HostPort)

	hostportlog.Info("default", "name", r.Name)

	if r.DeletionTimestamp.IsZero() {
		controllerutil.AddFinalizer(r, hostportv1alpha1.HostPortFinalizer)
	}

	return nil
}

var _ webhook.Validator = &HostPortWebhook{}

func (w *HostPortWebhook) ValidateCreate(obj runtime.Object) error {
	_ = context.Background()
	r := obj.(*hostportv1alpha1.HostPort)

	hostportlog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: hostportv1alpha1.GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)

}

func (w *HostPortWebhook) ValidateUpdate(obj runtime.Object, old runtime.Object) error {
	_ = context.Background()
	r := obj.(*hostportv1alpha1.HostPort)

	hostportlog.Info("validate update", "name", r.Name)
	oldHP := old.(*hostportv1alpha1.HostPort)

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
	if oldHP.Status.Port == 0 && r.Status.Port > 0 && r.Status.Phase != hostportv1alpha1.HostPortPhaseAllocated {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("status").Child("port"), r.Status.Port,
				fmt.Sprintf("port can only be set when also setting the phase to %s", hostportv1alpha1.HostPortPhaseAllocated)),
		)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: hostportv1alpha1.GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)

}

func (w *HostPortWebhook) ValidateDelete(obj runtime.Object) error {
	r := obj.(*hostportv1alpha1.HostPortClass)

	hostportlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
