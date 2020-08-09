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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hostportv1alpha1 "github.com/rmb938/hostport-allocator/api/v1alpha1"
)

// HostPortAllocationReconciler reconciles a HostPortAllocation object
type HostPortAllocationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hostport.rmb938.com,resources=hostportallocations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hostport.rmb938.com,resources=hostportallocations/status,verbs=get;update;patch

func (r *HostPortAllocationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("hostportallocation", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *HostPortAllocationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&hostportv1alpha1.HostPortAllocation{}, "spec.class", func(rawObj runtime.Object) []string {
		hpa := rawObj.(*hostportv1alpha1.HostPortAllocation)
		return []string{hpa.Spec.Class}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&hostportv1alpha1.HostPortAllocation{}).
		Complete(r)
}
