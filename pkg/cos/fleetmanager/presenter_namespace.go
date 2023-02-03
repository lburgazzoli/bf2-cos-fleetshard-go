package fleetmanager

import (
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/conditions"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/pointer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PresentConnectorNamespaceDeploymentStatus(ns corev1.Namespace) controlplane.ConnectorNamespaceDeploymentStatus {
	answer := controlplane.ConnectorNamespaceDeploymentStatus{
		Id:      ns.Labels[cosmeta.MetaNamespaceID],
		Phase:   controlplane.CONNECTORNAMESPACESTATE_READY,
		Version: ns.Labels[cosmeta.MetaNamespaceRevision],
	}

	switch ns.Status.Phase {
	case corev1.NamespaceActive:
		answer.Phase = controlplane.CONNECTORNAMESPACESTATE_READY
	default:
		answer.Phase = controlplane.CONNECTORNAMESPACESTATE_DISCONNECTED
	}

	return answer
}

func PresentConnectorDeploymentStatus(res cosv2.ManagedConnector) controlplane.ConnectorDeploymentStatus {
	answer := controlplane.ConnectorDeploymentStatus{}
	answer.ResourceVersion = &res.Spec.DeploymentResourceVersion
	answer.Conditions = make([]controlplane.MetaV1Condition, len(res.Status.Conditions))

	if c := conditions.Find(res.Status.Conditions, conditions.ConditionTypeProvisioned); c != nil {
		switch {
		case c.Reason == conditions.ConditionReasonStopping:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_DEPROVISIONING)
		case c.Reason == conditions.ConditionReasonStopped:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_STOPPED)
		case c.Reason == conditions.ConditionReasonDeleting:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_DELETING)
		case c.Reason == conditions.ConditionReasonDeleted:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_DELETED)
		case c.Status == metav1.ConditionFalse:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_PROVISIONING)
		}
	}

	if c := conditions.Find(res.Status.Conditions, conditions.ConditionTypeReady); c != nil {
		switch {
		case c.Status == metav1.ConditionTrue:
			answer.Phase = pointer.Of(controlplane.CONNECTORSTATE_READY)
		case c.Status == metav1.ConditionFalse && c.Reason == conditions.ConditionReasonError:
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
