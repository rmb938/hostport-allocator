package meta

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	intmetav1 "github.com/rmb938/hostport-allocator/apis/meta/v1"
)

// TODO: copied from https://github.com/kubernetes/apimachinery/blob/master/pkg/api/meta/conditions.go
//   convert to using apimachinery once 1.19 is stable and more widely used

// SetStatusCondition sets the corresponding condition in conditions to newCondition.
// conditions must be non-nil.
//  1. if the condition of the specified type already exists (all fields of the existing condition are updated to
//     newCondition, LastTransitionTime is set to now if the new status differs from the old status)
//  2. if a condition of the specified type does not exist (LastTransitionTime is set to now() if unset, and newCondition is appended)
func SetStatusCondition(conditions *[]intmetav1.Condition, newCondition intmetav1.Condition) {
	if conditions == nil {
		return
	}
	existingCondition := FindStatusCondition(*conditions, newCondition.Type)
	if existingCondition == nil {
		if newCondition.LastTransitionTime.IsZero() {
			newCondition.LastTransitionTime = metav1.NewTime(time.Now())
		}
		*conditions = append(*conditions, newCondition)
		return
	}

	if existingCondition.Status != newCondition.Status {
		existingCondition.Status = newCondition.Status
		if !newCondition.LastTransitionTime.IsZero() {
			existingCondition.LastTransitionTime = newCondition.LastTransitionTime
		} else {
			existingCondition.LastTransitionTime = metav1.NewTime(time.Now())
		}
	}

	existingCondition.Reason = newCondition.Reason
	existingCondition.Message = newCondition.Message
}

// RemoveStatusCondition removes the corresponding conditionType from conditions.
// conditions must be non-nil.
func RemoveStatusCondition(conditions *[]intmetav1.Condition, conditionType string) {
	if conditions == nil {
		return
	}
	newConditions := make([]intmetav1.Condition, 0, len(*conditions)-1)
	for _, condition := range *conditions {
		if condition.Type != conditionType {
			newConditions = append(newConditions, condition)
		}
	}

	*conditions = newConditions
}

// FindStatusCondition finds the conditionType in conditions.
func FindStatusCondition(conditions []intmetav1.Condition, conditionType string) *intmetav1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}

	return nil
}

// IsStatusConditionTrue returns true when the conditionType is present and set to `metav1.ConditionTrue`
func IsStatusConditionTrue(conditions []intmetav1.Condition, conditionType string) bool {
	return IsStatusConditionPresentAndEqual(conditions, conditionType, intmetav1.ConditionTrue)
}

// IsStatusConditionFalse returns true when the conditionType is present and set to `metav1.ConditionFalse`
func IsStatusConditionFalse(conditions []intmetav1.Condition, conditionType string) bool {
	return IsStatusConditionPresentAndEqual(conditions, conditionType, intmetav1.ConditionFalse)
}

// IsStatusConditionPresentAndEqual returns true when conditionType is present and equal to status.
func IsStatusConditionPresentAndEqual(conditions []intmetav1.Condition, conditionType string, status intmetav1.ConditionStatus) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status == status
		}
	}
	return false
}
