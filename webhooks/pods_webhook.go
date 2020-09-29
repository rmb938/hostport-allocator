package webhooks

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	hostportv1alpha1 "github.com/rmb938/hostport-allocator/api/v1alpha1"
	"github.com/rmb938/hostport-allocator/webhook"
	"github.com/rmb938/hostport-allocator/webhook/admission"
)

// log is for logging in this package.
var podlog = logf.Log.WithName("pod-resource")

type PodWebhook struct {
	client client.Client
}

func (w *PodWebhook) SetupWebhookWithManager(mgr ctrl.Manager) {
	w.client = mgr.GetClient()
	hookServer := mgr.GetWebhookServer()

	hookServer.Register("/mutate-v1-pod", admission.DefaultingWebhookFor(w, &corev1.Pod{}))
}

var _ webhook.Defaulter = &PodWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *PodWebhook) Default(obj runtime.Object) error {
	ctx := context.Background()
	r := obj.(*corev1.Pod)

	podlog.Info("default", "name", r.Name, "namespace", r.Namespace)

	var allErrs field.ErrorList

	definedClaims := make(map[string]string)
	for annotation, value := range r.Annotations {
		if strings.HasPrefix(annotation, hostportv1alpha1.HostPortPodAnnotationClaimPrefix+"/") {
			portName := strings.Split(annotation, "/")[1]

			if len(portName) == 0 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("metadata").Child("annotations").Child(annotation), annotation,
					"Annotation name must contain the port name"))
			}

			if len(value) == 0 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("metadata").Child("annotations").Child(annotation), value,
					"Annotation value must contain the claim name"))
			}

			definedClaims[portName] = value
		}
	}

	// host ports must not be set
	// if claims defined
	//  all ports must be named
	//  all ports must have unique names
	portNames := make(map[string]struct{ containerIndex, portIndex int })
	for containerIndex, container := range r.Spec.Containers {
		for portIndex, port := range container.Ports {
			path := field.NewPath("spec").Child("containers").Index(containerIndex).Child("ports").Index(portIndex).Child("name")
			if len(definedClaims) > 0 {
				if len(port.Name) == 0 {
					allErrs = append(allErrs, field.Invalid(path, port.Name,
						"Port name must be set"))
				}

				if _, ok := portNames[port.Name]; ok {
					allErrs = append(allErrs, field.Duplicate(path, port.Name))
				}
				portNames[port.Name] = struct{ containerIndex, portIndex int }{containerIndex: containerIndex, portIndex: portIndex}
			}

			if _, ok := definedClaims[port.Name]; !ok && port.HostPort > 0 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("containers").Index(containerIndex).Child("ports").Index(portIndex).Child("hostPort"), port.HostPort,
					"host ports cannot be set"))
			}
		}
	}

	for portName, claimName := range definedClaims {
		if len(portName) == 0 || len(claimName) == 0 {
			continue
		}

		path := field.NewPath("metadata").Child("annotations").Child(fmt.Sprintf("%s/%s", hostportv1alpha1.HostPortPodAnnotationClaimPrefix, portName))

		hpc := &hostportv1alpha1.HostPortClaim{}
		err := w.client.Get(ctx, types.NamespacedName{Namespace: r.Namespace, Name: claimName}, hpc)
		if err != nil {
			if apierrors.IsNotFound(err) {
				allErrs = append(allErrs, field.NotFound(path, claimName))
			} else {
				allErrs = append(allErrs, field.InternalError(path, err))
			}
			continue
		}

		// if not bound and not deleting don't allow
		// hpc can be deleting and still be usable (it won't go poof until all pods using it are gone)
		if hpc.Status.Phase != hostportv1alpha1.HostPortClaimPhaseBound && hpc.Status.Phase != hostportv1alpha1.HostPortClaimPhaseDeleting {
			allErrs = append(allErrs, field.Invalid(path, claimName,
				"hostPortClaim is not bound to a host port yet"))
			continue
		}

		// hpc doesn't have finalizer so it's about to go poof so we can't use it
		if controllerutil.ContainsFinalizer(hpc, hostportv1alpha1.HostPortFinalizer) == false {
			allErrs = append(allErrs, field.Invalid(path, claimName,
				"hostPortClaim is deleting"))
			continue
		}

		hp := &hostportv1alpha1.HostPort{}
		err = w.client.Get(ctx, types.NamespacedName{Name: hpc.Spec.HostPortName}, hp)
		if err != nil {
			if apierrors.IsNotFound(err) {
				allErrs = append(allErrs, field.NotFound(path.Child("hostPort"), hpc.Spec.HostPortName))
			} else {
				allErrs = append(allErrs, field.InternalError(path.Child("hostPort"), err))
			}
			continue
		}

		if hp.Status.Phase != hostportv1alpha1.HostPortPhaseAllocated {
			allErrs = append(allErrs, field.Invalid(path.Child("hostPort"), hp.Name,
				"hostport is not allocated a port yet"))
			continue
		}

		if hp.Status.Port == 0 {
			allErrs = append(allErrs, field.Invalid(path.Child("hostPort"), hp.Name,
				"hostport is not allocated a port yet"))
			continue
		}

		if portLocation, ok := portNames[portName]; ok {
			r.Annotations[hostportv1alpha1.HostPortPodAnnotationPortPrefix+"/"+portName] = strconv.Itoa(hp.Status.Port)
			r.Spec.Containers[portLocation.containerIndex].Ports[portLocation.portIndex].HostPort = int32(hp.Status.Port)
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "", Kind: r.Kind},
		r.Name, allErrs)
}
