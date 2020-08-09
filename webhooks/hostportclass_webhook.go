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

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *HostPortClassWebhook) Default(obj runtime.Object) {
	r := obj.(*hostportv1alpha1.HostPortClass)

	hostportclasslog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

var _ webhook.Validator = &HostPortClassWebhook{}

func (w *HostPortClassWebhook) validatePool(r *hostportv1alpha1.HostPortClass) field.ErrorList {
	var allErrs field.ErrorList

	for i, p1 := range r.Spec.Pool {
		name := p1.Name
		start := p1.Start
		end := p1.End
		if end < start {
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec").Child("pool").Index(i), end,
				"pool end must be greater than or equal to pool start"),
			)
		}

		dupNames := 0

		for _, p2 := range r.Spec.Pool {
			if name == p2.Name {
				dupNames += 1

				// this duplicate is us (or an exact copy of us) so don't do other checks
				if start == p2.Start && end == p2.End {
					continue
				}
			}

			if (start >= p2.Start && start <= p2.End) || end >= p2.Start && end <= p2.End {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("pool").Index(i), r.Spec.Pool[i],
					fmt.Sprintf("pool contains ports that are already defined in pool %s", p2.Name)),
				)
			}
		}

		// we will always find our self so look for > 1
		if dupNames > 1 {
			allErrs = append(allErrs, field.Duplicate(
				field.NewPath("spec").Child("pool").Index(i).Key("name"), name),
			)
		}
	}

	return allErrs
}

func (w *HostPortClassWebhook) ValidateCreate(obj runtime.Object) error {
	_ = context.Background()
	r := obj.(*hostportv1alpha1.HostPortClass)

	hostportclasslog.Info("validate create", "name", r.Name)

	allErrs := w.validatePool(r)

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

	allErrs := w.validatePool(r)

	// TODO: how do we handle changing pools
	//  i.e removing existing pools
	//  i.e modifying existing pools

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
