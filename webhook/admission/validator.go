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

package admission

import (
	"context"
	"net/http"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Validator interface {
	ValidateCreate(obj runtime.Object) error
	ValidateUpdate(obj runtime.Object, old runtime.Object) error
	ValidateDelete(obj runtime.Object) error
}

// ValidatingWebhookFor creates a new Webhook for validating the provided type.
func ValidatingWebhookFor(validator Validator, apiType runtime.Object) *admission.Webhook {
	return &admission.Webhook{
		Handler: &validatingHandler{
			apiType:   apiType,
			validator: validator,
		},
	}
}

type validatingHandler struct {
	apiType   runtime.Object
	validator Validator
	decoder   *admission.Decoder
}

var _ admission.DecoderInjector = &validatingHandler{}

// InjectDecoder injects the decoder into a validatingHandler.
func (h *validatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// Handle handles admission requests.
func (h *validatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if h.validator == nil {
		panic("validator should never be nil")
	}

	// Get the object in the request
	obj := h.apiType.DeepCopyObject().(runtime.Object)
	if req.Operation == v1beta1.Create {
		err := h.decoder.Decode(req, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = h.validator.ValidateCreate(obj)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	if req.Operation == v1beta1.Update {
		oldObj := obj.DeepCopyObject()

		err := h.decoder.DecodeRaw(req.Object, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		err = h.decoder.DecodeRaw(req.OldObject, oldObj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = h.validator.ValidateUpdate(obj, oldObj)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	if req.Operation == v1beta1.Delete {
		// In reference to PR: https://github.com/kubernetes/kubernetes/pull/76346
		// OldObject contains the object being deleted
		err := h.decoder.DecodeRaw(req.OldObject, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = h.validator.ValidateDelete(obj)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed("")
}
