package fleetmanager

import (
	"context"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
)

type Client interface {
	GetNamespaces(context.Context, int64) ([]controlplane.ConnectorNamespaceDeployment, error)
	GetConnectors(context.Context, int64) ([]controlplane.ConnectorDeployment, error)
	UpdateClusterStatus(ctx context.Context, status controlplane.ConnectorClusterStatus) error
	UpdateConnectorDeploymentStatus(ctx context.Context, id string, status controlplane.ConnectorDeploymentStatus) error
}

type ResourceGone struct {
	error string
	code  int
}

func (e ResourceGone) Error() string {
	return e.error
}

func (e ResourceGone) Code() int {
	return e.code
}
