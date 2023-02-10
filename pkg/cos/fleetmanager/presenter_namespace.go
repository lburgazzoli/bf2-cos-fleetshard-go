package fleetmanager

import (
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	corev1 "k8s.io/api/core/v1"
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
