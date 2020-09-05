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
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// HostPortClaimReconciler reconciles a HostPortClaim object
type HostPortClaimReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hostport.rmb938.com,resources=hostportclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hostport.rmb938.com,resources=hostportclaims/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

func (r *HostPortClaimReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("hostportclaim", req.NamespacedName)

	hpc := &hostportv1alpha1.HostPortClaim{}
	err := r.Get(ctx, req.NamespacedName, hpc)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return ctrl.Result{}, err
	}

	if hpc.DeletionTimestamp.IsZero() == false {
		// is deleting but the phase isn't deleting so set it
		if hpc.Status.Phase != hostportv1alpha1.HostPortClaimPhaseDeleting {
			hpc.Status.Phase = hostportv1alpha1.HostPortClaimPhaseDeleting
			err = r.Status().Update(ctx, hpc)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		if controllerutil.ContainsFinalizer(hpc, hostportv1alpha1.HostPortFinalizer) {
			return ctrl.Result{}, nil
		}

		// don't allow deletion when in use
		podList := &corev1.PodList{}
		err := r.List(ctx, podList, client.InNamespace(hpc.Namespace))
		if err != nil {
			return ctrl.Result{}, err
		}

		if len(podList.Items) > 0 {
			for _, pod := range podList.Items {
				if pod.Annotations == nil {
					continue
				}

				for annotation, value := range pod.Annotations {
					if strings.HasPrefix(annotation, hostportv1alpha1.HostPortPodAnnotationClaimPrefix+"/") {
						if value == hpc.Name {
							return ctrl.Result{}, nil
						}
					}
				}
			}
		}

		// remove the finalizer
		controllerutil.RemoveFinalizer(hpc, hostportv1alpha1.HostPortFinalizer)

		// send update to remove finalizer
		err = r.Update(ctx, hpc)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if len(hpc.Status.Phase) == 0 {
		hpc.Status.Phase = hostportv1alpha1.HostPortClaimPhasePending
		err = r.Status().Update(ctx, hpc)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if hpc.Status.Phase == hostportv1alpha1.HostPortClaimPhasePending {

		hp := &hostportv1alpha1.HostPort{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("hpc-%s", hpc.UID),
			},
			Spec: hostportv1alpha1.HostPortSpec{
				ClaimRef: &corev1.ObjectReference{
					Namespace: hpc.Namespace,
					Name:      hpc.Name,
					UID:       hpc.UID,
				},
				HostPortClassName: hpc.Spec.HostPortClassName,
			},
		}

		err := r.Create(ctx, hp)
		if err != nil {
			if apierrors.IsAlreadyExists(err) == false {
				return ctrl.Result{}, err
			}
		}

		hpc.Status.HostPortName = fmt.Sprintf("hpc-%s", hpc.UID)
		hpc.Status.Phase = hostportv1alpha1.HostPortClaimPhaseBound
		err = r.Status().Update(ctx, hpc)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *HostPortClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hostportv1alpha1.HostPortClaim{}).
		Watches(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			pod := a.Object.(*corev1.Pod)
			var req []reconcile.Request

			for annotation, value := range pod.Annotations {
				if strings.HasPrefix(annotation, hostportv1alpha1.HostPortPodAnnotationClaimPrefix+"/") {
					req = append(req, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Namespace: pod.Namespace,
							Name:      value,
						},
					})
				}
			}

			return req
		})}).
		Watches(&source.Kind{Type: &hostportv1alpha1.HostPort{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			hp := a.Object.(*hostportv1alpha1.HostPort)
			var req []reconcile.Request

			if hp.Spec.ClaimRef != nil {
				req = append(req, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: hp.Spec.ClaimRef.Namespace,
						Name:      hp.Spec.ClaimRef.Name,
					},
				})
			}

			return req
		})}).
		Complete(r)
}
