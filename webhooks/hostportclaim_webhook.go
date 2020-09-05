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
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	hostportv1alpha1 "github.com/rmb938/hostport-allocator/api/v1alpha1"
	"github.com/rmb938/hostport-allocator/webhook"
	"github.com/rmb938/hostport-allocator/webhook/admission"
)

// log is for logging in this package.
var hostportclaimlog = logf.Log.WithName("hostportclaim-resource")

type HostPortClaimWebhook struct {
	client client.Client
}

func (w *HostPortClaimWebhook) SetupWebhookWithManager(mgr ctrl.Manager) {
	w.client = mgr.GetClient()
	hookServer := mgr.GetWebhookServer()

	hookServer.Register("/mutate-hostport-rmb938-com-v1alpha1-hostportclaim", admission.DefaultingWebhookFor(w, &hostportv1alpha1.HostPortClaim{}))
	hookServer.Register("/validate-hostport-rmb938-com-v1alpha1-hostportclaim", admission.ValidatingWebhookFor(w, &hostportv1alpha1.HostPortClaim{}))
}

var _ webhook.Defaulter = &HostPortClaimWebhook{}

func (w *HostPortClaimWebhook) Default(obj runtime.Object) error {
	r := obj.(*hostportv1alpha1.HostPortClaim)

	hostportclaimlog.Info("default", "name", r.Name)

	if r.DeletionTimestamp.IsZero() {
		hasFinalizer := false

		for _, finalizer := range r.Finalizers {
			if finalizer == hostportv1alpha1.HostPortFinalizer {
				hasFinalizer = true
				break
			}
		}

		if hasFinalizer == false {
			r.Finalizers = append(r.Finalizers, hostportv1alpha1.HostPortFinalizer)
		}
	}

	return nil
}

var _ webhook.Validator = &HostPortClaimWebhook{}

func (w *HostPortClaimWebhook) ValidateCreate(obj runtime.Object) error {
	_ = context.Background()
	r := obj.(*hostportv1alpha1.HostPortClaim)

	hostportclaimlog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: hostportv1alpha1.GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)

}

func (w *HostPortClaimWebhook) ValidateUpdate(obj runtime.Object, old runtime.Object) error {
	_ = context.Background()
	r := obj.(*hostportv1alpha1.HostPortClaim)

	hostportclaimlog.Info("validate update", "name", r.Name)
	oldHPC := old.(*hostportv1alpha1.HostPortClaim)

	var allErrs field.ErrorList

	if r.Spec.HostPortClassName != oldHPC.Spec.HostPortClassName {
		allErrs = append(allErrs,
			field.Forbidden(field.NewPath("spec").Child("hostPortClassName"),
				"cannot change hostPortClassName"),
		)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: hostportv1alpha1.GroupVersion.Group, Kind: r.Kind},
		r.Name, allErrs)

}

func (w *HostPortClaimWebhook) ValidateDelete(obj runtime.Object) error {
	r := obj.(*hostportv1alpha1.HostPortClaim)

	hostportclaimlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
