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
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	hostportv1alpha1 "github.com/rmb938/hostport-allocator/api/v1alpha1"
)

// HostPortReconciler reconciles a HostPort object
type HostPortReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	allocationLock sync.Mutex
}

// +kubebuilder:rbac:groups=hostport.rmb938.com,resources=hostports,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hostport.rmb938.com,resources=hostports/status,verbs=get;update;patch

func (r *HostPortReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("hostport", req.NamespacedName)

	hp := &hostportv1alpha1.HostPort{}
	err := r.Get(ctx, req.NamespacedName, hp)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// if port is deleting
	if hp.DeletionTimestamp.IsZero() == false {

		// is deleting but the phase isn't deleting so set it
		if hp.Status.Phase != hostportv1alpha1.HostPortPhaseDeleting {
			hp.Status.Phase = hostportv1alpha1.HostPortPhaseDeleting
			err = r.Status().Update(ctx, hp)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		if controllerutil.ContainsFinalizer(hp, hostportv1alpha1.HostPortFinalizer) == false {
			return ctrl.Result{}, nil
		}

		// don't allow deletion when in use
		if hp.Spec.ClaimRef != nil {
			hpc := &hostportv1alpha1.HostPortClaim{}
			err := r.Get(ctx, types.NamespacedName{Namespace: hp.Spec.ClaimRef.Namespace, Name: hp.Spec.ClaimRef.Name}, hpc)
			if err != nil {
				if apierrors.IsNotFound(err) == false {
					return ctrl.Result{}, err
				}
				hpc = nil
			}

			if hpc != nil {
				if hpc.UID == hp.Spec.ClaimRef.UID {
					// can't delete because claimref hpc exists
					return ctrl.Result{}, nil
				}
			}
		}

		// remove the finalizer
		controllerutil.RemoveFinalizer(hp, hostportv1alpha1.HostPortFinalizer)

		// send update to remove finalizer
		err = r.Update(ctx, hp)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if len(hp.Status.Phase) == 0 {
		hp.Status.Phase = hostportv1alpha1.HostPortPhasePending
		err = r.Status().Update(ctx, hp)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if hp.Status.Phase == hostportv1alpha1.HostPortPhasePending {

		hpcl := &hostportv1alpha1.HostPortClass{}
		err := r.Get(ctx, types.NamespacedName{Name: hp.Spec.HostPortClassName}, hpcl)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// TODO: event saying can't find host port class
			}

			return ctrl.Result{}, err
		}

		r.allocationLock.Lock()
		defer r.allocationLock.Unlock()

		// get all ports
		hostPortList := &hostportv1alpha1.HostPortList{}
		err = r.List(ctx, hostPortList)
		if err != nil {
			return ctrl.Result{}, err
		}

		usedPorts := make(map[int]struct{})
		for _, i := range hostPortList.Items {
			if i.Status.Port > 0 {
				usedPorts[i.Status.Port] = struct{}{}
			}
		}

		availablePorts := make([]int, 0)
		for _, pool := range hpcl.Spec.Pools {
			for port := pool.Start; port <= pool.End; port++ {
				if _, ok := usedPorts[port]; !ok {
					availablePorts = append(availablePorts, port)
				}
			}
		}

		if len(availablePorts) == 0 {
			// TODO: event saying can't find any ports

			err := fmt.Errorf("no free ports to allocate")
			return ctrl.Result{}, err
		}

		hp.Status.Port = availablePorts[0]
		hp.Status.Phase = hostportv1alpha1.HostPortPhaseAllocated
		err = r.Status().Update(ctx, hp)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if hp.Status.Phase == hostportv1alpha1.HostPortPhaseAllocated {
		// TODO: reclaimPolicy
		//  Retain - default for manually created HostPorts
		//    Do not delete when claimRef is gone, just set it to nil
		//  Delete - default for dynamically provisioned HostPorts
		//    Delete when ClaimRef is gone
		if hp.Spec.ClaimRef != nil {
			hpc := &hostportv1alpha1.HostPortClaim{}
			err := r.Get(ctx, types.NamespacedName{Namespace: hp.Spec.ClaimRef.Namespace, Name: hp.Spec.ClaimRef.Name}, hpc)
			if err != nil {
				if apierrors.IsNotFound(err) == false {
					return ctrl.Result{}, err
				}
				hpc = nil
			}

			// my hpc is gone or different so delete me
			if hpc == nil || hpc.UID != hp.Spec.ClaimRef.UID {
				err := r.Delete(ctx, hp)
				if err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *HostPortReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &hostportv1alpha1.HostPort{}, "spec.hostPortClassName", func(rawObj runtime.Object) []string {
		hp := rawObj.(*hostportv1alpha1.HostPort)
		return []string{hp.Spec.HostPortClassName}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&hostportv1alpha1.HostPort{}).
		Watches(&source.Kind{Type: &hostportv1alpha1.HostPortClaim{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			hpc := a.Object.(*hostportv1alpha1.HostPortClaim)
			var req []reconcile.Request

			if len(hpc.Spec.HostPortName) > 0 {
				req = append(req, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name: hpc.Spec.HostPortName,
					},
				})
			}

			return req
		})}).
		Complete(r)
}
