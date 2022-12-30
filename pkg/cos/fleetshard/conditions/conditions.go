package conditions

import (
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const ConditionTypeReady string = "Ready"

const ConditionMessageProvisioned string = "Provisioned"
const ConditionReasonProvisioning string = "Provisioning"
const ConditionReasonProvisioned string = "Provisioned"
const ConditionMessageStopped string = "Stopped"
const ConditionReasonStopped string = "Stopped"
const ConditionMessageStopping string = "Stopping"
const ConditionReasonStopping string = "Stopping"
const ConditionMessageDeleted string = "Deleted"
const ConditionReasonDeleted string = "Deleted"
const ConditionMessageDeleting string = "Deleting"
const ConditionReasonDeleting string = "Deleting"
const ConditionMessageUnknown string = "Unknown"
const ConditionReasonUnknown string = "Unknown"
const ConditionMessageProvisioning string = "Provisioning"

func Ready(connector cosv2.ManagedConnector) cosv2.Condition {
	ready := cosv2.Condition{
		Condition: v1.Condition{
			Type:               ConditionTypeReady,
			Status:             v1.ConditionFalse,
			Reason:             ConditionReasonUnknown,
			Message:            ConditionMessageUnknown,
			ObservedGeneration: connector.Status.ObservedGeneration,
		},
		ResourceRevision: connector.Spec.Deployment.DeploymentResourceVersion,
	}

	if connector.Generation != connector.Status.ObservedGeneration {
		ready.Reason = ConditionMessageProvisioning
		ready.Message = ConditionReasonProvisioning
	}

	return ready
}

func UpdateReady(connector *cosv2.ManagedConnector, status v1.ConditionStatus, reason string, message string) {
	Update(&connector.Status.Conditions, ConditionTypeReady, func(condition *cosv2.Condition) {
		condition.Status = status
		condition.Reason = reason
		condition.Message = message
	})
}

func Update(conditions *[]cosv2.Condition, conditionType string, consumer func(*cosv2.Condition)) {
	c := Find(*conditions, conditionType)
	if c == nil {
		c = &cosv2.Condition{
			Condition: v1.Condition{
				Type: conditionType,
			},
		}
	}

	consumer(c)

	Set(conditions, *c)

}

// Find finds the conditionType in conditions.
func Find(conditions []cosv2.Condition, conditionType string) *cosv2.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}

	return nil
}

func Set(conditions *[]cosv2.Condition, newCondition cosv2.Condition) {
	if conditions == nil {
		return
	}
	existingCondition := Find(*conditions, newCondition.Type)
	if existingCondition == nil {
		if newCondition.LastTransitionTime.IsZero() {
			newCondition.LastTransitionTime = v1.NewTime(time.Now())
		}
		*conditions = append(*conditions, newCondition)
		return
	}

	if existingCondition.Status != newCondition.Status {
		existingCondition.Status = newCondition.Status
		if !newCondition.LastTransitionTime.IsZero() {
			existingCondition.LastTransitionTime = newCondition.LastTransitionTime
		} else {
			existingCondition.LastTransitionTime = v1.NewTime(time.Now())
		}
	}

	existingCondition.Reason = newCondition.Reason
	existingCondition.Message = newCondition.Message
	existingCondition.ObservedGeneration = newCondition.ObservedGeneration
}
