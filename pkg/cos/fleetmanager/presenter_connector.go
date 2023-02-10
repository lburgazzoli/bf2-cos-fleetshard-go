package fleetmanager

import (
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/conditions"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/pointer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PresentConnectorDeploymentStatus(res cosv2.ManagedConnector) controlplane.ConnectorDeploymentStatus {
	answer := controlplane.ConnectorDeploymentStatus{
		ResourceVersion: pointer.Of(res.Status.ObservedDeploymentResourceVersion),
		Conditions:      make([]controlplane.MetaV1Condition, len(res.Status.Conditions)),
	}

	if p := conditions.Find(res.Status.Conditions, conditions.ConditionTypeProvisioned); p != nil {
		switch {

		// operator has not yet taken into account the new revision
		case p.ResourceRevision != res.Spec.DeploymentResourceVersion && res.Spec.DesiredState == cosv2.DesiredStateReady:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_PROVISIONING)
		case p.ResourceRevision != res.Spec.DeploymentResourceVersion && res.Spec.DesiredState == cosv2.DesiredStateStopped:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_DEPROVISIONING)
		case p.ResourceRevision != res.Spec.DeploymentResourceVersion && res.Spec.DesiredState == cosv2.DesiredStateDeleted:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_DEPROVISIONING)

		// operator has provisioned the operand resources
		case p.Reason == conditions.ConditionReasonStopping:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_DEPROVISIONING)
		case p.Reason == conditions.ConditionReasonStopped:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_STOPPED)
		case p.Reason == conditions.ConditionReasonDeleting:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_DELETING)
		case p.Reason == conditions.ConditionReasonDeleted:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_DELETED)
		case p.Status == metav1.ConditionFalse:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_PROVISIONING)

		}
	}

	if r := conditions.Find(res.Status.Conditions, conditions.ConditionTypeReady); r != nil {
		switch {

		// operator has not yet provisioned the desired state
		case r.ResourceRevision != res.Spec.DeploymentResourceVersion && res.Spec.DesiredState == cosv2.DesiredStateReady:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_PROVISIONING)
		case r.ResourceRevision != res.Spec.DeploymentResourceVersion && res.Spec.DesiredState == cosv2.DesiredStateStopped:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_DEPROVISIONING)
		case r.ResourceRevision != res.Spec.DeploymentResourceVersion && res.Spec.DesiredState == cosv2.DesiredStateDeleted:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_DEPROVISIONING)

		// operator has provisioned the operand resources
		case r.Status == metav1.ConditionTrue:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_READY)
		case r.Status == metav1.ConditionFalse && r.Reason == conditions.ConditionReasonError:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_FAILED)

		}
	}

	for i := range res.Status.Conditions {
		answer.Conditions[i] = controlplane.MetaV1Condition{
			Type:    pointer.Of(res.Status.Conditions[i].Type),
			Status:  pointer.Of(string(res.Status.Conditions[i].Status)),
			Reason:  pointer.Of(res.Status.Conditions[i].Reason),
			Message: pointer.Of(res.Status.Conditions[i].Message),
		}
	}

	return answer
}
