package v1

import (
	"fmt"
)

type ConditionType string

// +kubebuilder:validation:Enum=True;False;Error;Unknown
type ConditionStatus string

const (
	ConditionStatusTrue    ConditionStatus = "True"
	ConditionStatusFalse   ConditionStatus = "False"
	ConditionStatusError   ConditionStatus = "Error"
	ConditionStatusUnknown ConditionStatus = "Unknown"
)

type ConditionObject interface {
	GetConditions() []StatusCondition
	SetConditions(conditions []StatusCondition)
	GetCondition(conditionType ConditionType) *StatusCondition
	SetCondition(newCondition *StatusCondition) error
}

func (scs *StatusConditions) GetConditions() []StatusCondition {
	return scs.Conditions
}

func (scs *StatusConditions) GetCondition(conditionType ConditionType) *StatusCondition {
	for _, cond := range scs.Conditions {
		if cond.Type == conditionType {
			return &cond
		}
	}

	return nil
}

func (scs *StatusConditions) SetConditions(conditions []StatusCondition) {
	scs.Conditions = conditions
}

func (scs *StatusConditions) SetCondition(newCondition *StatusCondition) error {
	if len(newCondition.Type) == 0 || len(newCondition.Status) == 0 || newCondition.LastTransitionTime.IsZero() {
		return fmt.Errorf("condition is not fully formed")
	}

	// Search through existing conditions
	for idx, cond := range scs.Conditions {
		// Skip unrelated conditions
		if cond.Type != newCondition.Type {
			continue
		}

		// If this update doesn't contain a state transition, we don't update
		// the conditions LastTransitionTime to Now()
		if cond.Status == newCondition.Status {
			newCondition.LastTransitionTime = cond.LastTransitionTime
		}

		// Overwrite the existing condition
		scs.Conditions[idx] = *newCondition
		return nil
	}

	scs.Conditions = append(scs.Conditions, *newCondition)
	return nil
}
