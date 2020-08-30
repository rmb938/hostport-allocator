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
	"sync"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
		err = client.IgnoreNotFound(err)
		return ctrl.Result{}, err
	}

	// if port is deleting
	if hp.DeletionTimestamp.IsZero() == false {

		// class is deleting but the phase isn't deleting so set it
		if hp.Status.Phase != hostportv1alpha1.HostPortPhaseDeleting {
			hp.Status.Phase = hostportv1alpha1.HostPortPhaseDeleting
			err = r.Status().Update(ctx, hp)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		hasFinalizer := false

		for _, finalizer := range hp.Finalizers {
			if finalizer == hostportv1alpha1.HostPortFinalizer {
				hasFinalizer = true
				break
			}
		}

		// It doesn't have finalizer to ignore it
		if hasFinalizer == false {
			return ctrl.Result{}, nil
		}

		// remove the finalizer to delete the pool
		for i, finalizer := range hp.Finalizers {
			if finalizer == hostportv1alpha1.HostPortFinalizer {
				hp.Finalizers = append(hp.Finalizers[:i], hp.Finalizers[i+1:]...)
				break
			}
		}

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

		hostPortPools := &hostportv1alpha1.HostPortPoolList{}
		err = r.List(ctx, hostPortPools, client.MatchingFields{"spec.hostPortClassName": hpcl.Name})
		if err != nil {
			return ctrl.Result{}, err
		}

		if len(hostPortPools.Items) == 0 {
			// TODO: event saying can't find any pools for class
			return ctrl.Result{Requeue: true}, nil
		}

		portMap := make(map[string][]int)

		// generate port map
		for _, pool := range hostPortPools.Items {
			if pool.Status.Phase != hostportv1alpha1.HostPortPoolPhaseReady {
				continue
			}

			if pool.Spec.Enabled == false {
				continue
			}

			hostPorts := &hostportv1alpha1.HostPortList{}
			err := r.List(ctx, hostPorts, client.MatchingFields{"spec.hostPortPoolName": pool.Name})
			if err != nil {
				return ctrl.Result{}, err
			}

			ports := make([]int, 0)

		PoolNumbers:
			for port := pool.Spec.Start; port <= pool.Spec.End; port++ {
				for _, hostPort := range hostPorts.Items {
					if *hostPort.Status.Port == int64(port) {
						continue PoolNumbers
					}
				}

				ports = append(ports, port)
			}

			portMap[pool.Name] = ports
		}

		var foundPort *int64
		var foundPool *string

		for poolName, ports := range portMap {
			if len(ports) == 0 {
				continue
			}

			foundPort = pointer.Int64Ptr(int64(ports[0]))
			foundPool = &poolName
			break
		}

		// TODO: if we are in the middle of an allocation and a HPP turns to deleted, then we set the pool name and port
		//  how do we prevent this?

		hp.Status.Port = foundPort
		hp.Status.HostPortPoolName = foundPool
		hp.Status.Phase = hostportv1alpha1.HostPortPhaseAllocated
		err = r.Status().Update(ctx, hp)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *HostPortReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&hostportv1alpha1.HostPort{}, "spec.hostPortClassName", func(rawObj runtime.Object) []string {
		hp := rawObj.(*hostportv1alpha1.HostPort)
		return []string{hp.Spec.HostPortClassName}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(&hostportv1alpha1.HostPort{}, "spec.hostPortPoolName", func(rawObj runtime.Object) []string {
		hp := rawObj.(*hostportv1alpha1.HostPort)

		if hp.Status.HostPortPoolName == nil {
			return []string{}
		}

		return []string{*hp.Status.HostPortPoolName}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&hostportv1alpha1.HostPort{}).
		Complete(r)
}
