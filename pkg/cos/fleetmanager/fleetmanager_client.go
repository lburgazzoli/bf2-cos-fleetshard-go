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

type GenericError struct {
	Reason      string `json:"reason"`
	OperationID string `json:"operation_id"`
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Href        string `json:"href"`
	Code        string `json:"code"`
}

func (e GenericError) Error() string {
	return e.Reason
}
