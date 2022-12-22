package conditions

import (
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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

func Ready(connector v2.ManagedConnector) v1.Condition {
	ready := v1.Condition{
		Type:               ConditionTypeReady,
		Status:             v1.ConditionFalse,
		Reason:             ConditionReasonUnknown,
		Message:            ConditionMessageUnknown,
		ObservedGeneration: connector.Spec.Deployment.DeploymentResourceVersion,
	}

	if connector.Generation != connector.Status.ObservedGeneration {
		ready.Reason = ConditionMessageProvisioning
		ready.Message = ConditionReasonProvisioning
	}

	return ready
}

func SetReady(connector *v2.ManagedConnector, status v1.ConditionStatus, reason string, message string) {
	controller.UpdateStatusCondition(&connector.Status.Conditions, ConditionTypeReady, func(condition *v1.Condition) {
		condition.Status = status
		condition.Reason = reason
		condition.Message = message
	})
}
