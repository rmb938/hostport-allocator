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
	"k8s.io/apimachinery/pkg/runtime"
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

	// TODO(user): fill in your defaulting logic.
}

var _ webhook.Validator = &HostPortPoolWebhook{}

func (w *HostPortPoolWebhook) ValidateCreate(obj runtime.Object) error {
	r := obj.(*hostportv1alpha1.HostPortPool)

	hostportpoollog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

func (w *HostPortPoolWebhook) ValidateUpdate(obj runtime.Object, old runtime.Object) error {
	r := obj.(*hostportv1alpha1.HostPortPool)

	hostportpoollog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

func (w *HostPortPoolWebhook) ValidateDelete(obj runtime.Object) error {
	r := obj.(*hostportv1alpha1.HostPortPool)

	hostportpoollog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
