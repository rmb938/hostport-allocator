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

	"github.com/rmb938/hostport-allocator/api/meta"
	hostportv1alpha1 "github.com/rmb938/hostport-allocator/api/v1alpha1"
	intmetav1 "github.com/rmb938/hostport-allocator/apis/meta/v1"
)

// HostPortPoolReconciler reconciles a HostPortPool object
type HostPortPoolReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hostport.rmb938.com,resources=hostportpools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hostport.rmb938.com,resources=hostportpools/status,verbs=get;update;patch

func (r *HostPortPoolReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("hostportpool", req.NamespacedName)

	hpp := &hostportv1alpha1.HostPortPool{}
	err := r.Get(ctx, req.NamespacedName, hpp)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return ctrl.Result{}, err
	}

	// If pool is deleting
	if hpp.DeletionTimestamp.IsZero() == false {

		// pool is deleting but the phase isn't deleting so set it
		if hpp.Status.Phase != hostportv1alpha1.HostPortPoolPhaseDeleting {
			hpp.Status.Phase = hostportv1alpha1.HostPortPoolPhaseDeleting
			err = r.Status().Update(ctx, hpp)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		hasFinalizer := false

		for _, finalizer := range hpp.Finalizers {
			if finalizer == hostportv1alpha1.HostPortPoolFinalizer {
				hasFinalizer = true
				break
			}
		}

		// It doesn't have finalizer to ignore it
		if hasFinalizer == false {
			return ctrl.Result{}, nil
		}

		hostPorts := &hostportv1alpha1.HostPortList{}
		err := r.List(ctx, hostPorts, client.MatchingFields{"spec.hostPortClassName": hpp.Spec.HostPortClassName})
		if err != nil {
			return ctrl.Result{}, err
		}

		// Ports exist for this pool so we can't delete it yet
		if len(hostPorts.Items) > 0 {
			// TODO: event saying can't delete due to existing host ports
			// Has finalizer but host ports are still using this, we will auto re-reconcile when ports do
			return ctrl.Result{}, nil
		}

		// remove the finalizer to delete the pool
		for i, finalizer := range hpp.Finalizers {
			if finalizer == hostportv1alpha1.HostPortPoolFinalizer {
				hpp.Finalizers = append(hpp.Finalizers[:i], hpp.Finalizers[i+1:]...)
				break
			}
		}

		// send update to remove finalizer
		err = r.Update(ctx, hpp)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// pool has no phase so set one
	if len(hpp.Status.Phase) == 0 {
		hpp.Status.Phase = hostportv1alpha1.HostPortPoolPhasePending
		err = r.Status().Update(ctx, hpp)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// pool is pending
	if hpp.Status.Phase == hostportv1alpha1.HostPortPoolPhasePending {

		// check if has condition, if it doesn't add it
		condition := meta.FindStatusCondition(hpp.Status.Conditions, hostportv1alpha1.HostPortPoolConditionOverlap)
		if condition == nil {
			meta.SetStatusCondition(&hpp.Status.Conditions, intmetav1.Condition{
				Type:    hostportv1alpha1.HostPortPoolConditionOverlap,
				Status:  intmetav1.ConditionUnknown,
				Reason:  hostportv1alpha1.HostPortPoolConditionOverlapReasonNotChecked,
				Message: "",
			})

			err = r.Status().Update(ctx, hpp)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		// if condition is false we errored before so don't continue
		if meta.IsStatusConditionFalse(hpp.Status.Conditions, hostportv1alpha1.HostPortPoolConditionOverlap) {
			return ctrl.Result{}, nil
		}

		// get all pools
		hppList := &hostportv1alpha1.HostPortPoolList{}
		err := r.List(ctx, hppList)
		if err != nil {
			return ctrl.Result{}, err
		}

		foundOverlap := false

		// loop through pools
		for _, item := range hppList.Items {
			// ignore if it's us
			if item.UID == hpp.UID {
				continue
			}

			if hpp.Spec.Start >= item.Spec.Start && hpp.Spec.Start <= item.Spec.End {
				foundOverlap = true
				break
			}

			if hpp.Spec.End >= item.Spec.Start && hpp.Spec.End <= item.Spec.End {
				foundOverlap = true
				break
			}
		}

		// we found an overlap so don't continue
		if foundOverlap {
			// TODO: event with overlap

			// set condition to false with reason
			meta.SetStatusCondition(&hpp.Status.Conditions, intmetav1.Condition{
				Type:    hostportv1alpha1.HostPortPoolConditionOverlap,
				Status:  intmetav1.ConditionFalse,
				Reason:  hostportv1alpha1.HostPortPoolConditionOverlapReasonOverlap,
				Message: hostportv1alpha1.HostPortPoolConditionOverlapMessageOverlap,
			})

			err = r.Status().Update(ctx, hpp)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		// no overlap so set condition to true and phase to ready
		meta.SetStatusCondition(&hpp.Status.Conditions, intmetav1.Condition{
			Type:    hostportv1alpha1.HostPortPoolConditionOverlap,
			Status:  intmetav1.ConditionTrue,
			Reason:  hostportv1alpha1.HostPortPoolConditionOverlapReasonNoOverlap,
			Message: "",
		})

		hpp.Status.Phase = hostportv1alpha1.HostPortPoolPhaseReady
		err = r.Status().Update(ctx, hpp)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *HostPortPoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&hostportv1alpha1.HostPortPool{}, "spec.hostPortClassName", func(rawObj runtime.Object) []string {
		hpp := rawObj.(*hostportv1alpha1.HostPortPool)
		return []string{hpp.Spec.HostPortClassName}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&hostportv1alpha1.HostPortPool{}).
		Complete(r)
}
