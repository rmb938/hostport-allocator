package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type StatusConditions struct {
	// Conditions for the object
	// +kubebuilder:validation:Optional
	Conditions []StatusCondition `json:"conditions,omitempty"`
}

type StatusCondition struct {
	// Type of the condition
	// +kubebuilder:validation:Required
	Type ConditionType `json:"type"`

	// Status of the condition
	// +kubebuilder:validation:Required
	Status ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status change of this condition.
	// +kubebuilder:validation:Optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief machine readable explanation for the condition's last transition.
	// +kubebuilder:validation:Optional
	Reason string `json:"reason,omitempty"`

	// Message is a human readable description of the details of the last transition, complementing reason.
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`
}
