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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	hostportv1alpha1 "github.com/rmb938/hostport-allocator/api/v1alpha1"
)

// HostPortClassReconciler reconciles a HostPortClass object
type HostPortClassReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hostport.rmb938.com,resources=hostportclasses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hostport.rmb938.com,resources=hostportclasses/status,verbs=get;update;patch

func (r *HostPortClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("hostportclass", req.NamespacedName)

	hpcl := &hostportv1alpha1.HostPortClass{}
	err := r.Get(ctx, req.NamespacedName, hpcl)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return ctrl.Result{}, err
	}

	// If class is deleting
	if hpcl.DeletionTimestamp.IsZero() == false {
		if controllerutil.ContainsFinalizer(hpcl, hostportv1alpha1.HostPortFinalizer) == false {
			return ctrl.Result{}, nil
		}

		// remove the finalizer
		controllerutil.RemoveFinalizer(hpcl, hostportv1alpha1.HostPortFinalizer)

		// send update to remove finalizer
		err = r.Update(ctx, hpcl)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *HostPortClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hostportv1alpha1.HostPortClass{}).
		Complete(r)
}
