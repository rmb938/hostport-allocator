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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

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

func (r *HostPortClassReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("hostportclass", req.NamespacedName)

	hpcl := &hostportv1alpha1.HostPortClass{}
	err := r.Get(ctx, req.NamespacedName, hpcl)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return ctrl.Result{}, err
	}

	// If class is deleting
	if hpcl.DeletionTimestamp.IsZero() == false {

		// class is deleting but the phase isn't deleting so set it
		if hpcl.Status.Phase != hostportv1alpha1.HostPortClassPhaseDeleting {
			hpcl.Status.Phase = hostportv1alpha1.HostPortClassPhaseDeleting
			err = r.Status().Update(ctx, hpcl)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		hasFinalizer := false

		for _, finalizer := range hpcl.Finalizers {
			if finalizer == hostportv1alpha1.HostPortClassFinalizer {
				hasFinalizer = true
				break
			}
		}

		// It doesn't have finalizer to ignore it
		if hasFinalizer == false {
			return ctrl.Result{}, nil
		}

		hostPortPools := &hostportv1alpha1.HostPortPoolList{}
		err := r.List(ctx, hostPortPools, client.MatchingFields{"spec.hostPortClassName": hpcl.Name})
		if err != nil {
			return ctrl.Result{}, err
		}

		// Pools exist for this class so we can't delete it yet
		if len(hostPortPools.Items) > 0 {
			// TODO: event saying can't delete due to existing pools
			// Has finalizer but pools are still using this, we will auto re-reconcile when pools do
			return ctrl.Result{}, nil
		}

		// remove the finalizer to delete the pool
		for i, finalizer := range hpcl.Finalizers {
			if finalizer == hostportv1alpha1.HostPortClassFinalizer {
				hpcl.Finalizers = append(hpcl.Finalizers[:i], hpcl.Finalizers[i+1:]...)
				break
			}
		}

		// send update to remove finalizer
		err = r.Update(ctx, hpcl)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if len(hpcl.Status.Phase) == 0 {
		hpcl.Status.Phase = hostportv1alpha1.HostPortClassPhasePending
		err = r.Status().Update(ctx, hpcl)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if hpcl.Status.Phase == hostportv1alpha1.HostPortClassPhasePending {
		hpcl.Status.Phase = hostportv1alpha1.HostPortClassPhaseReady
		err = r.Status().Update(ctx, hpcl)
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
		Watches(&source.Kind{Type: &hostportv1alpha1.HostPortPool{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			hpp := a.Object.(*hostportv1alpha1.HostPortPool)
			var req []reconcile.Request

			req = append(req, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: hpp.Spec.HostPortClassName,
				},
			})

			return req
		})}).
		Complete(r)
}
