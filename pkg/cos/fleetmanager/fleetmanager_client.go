package fleetmanager

import (
	"context"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
)

type Client interface {
	GetNamespaces(context.Context, int64) ([]controlplane.ConnectorNamespaceDeployment, error)
	GetConnectors(context.Context, int64) ([]controlplane.ConnectorDeployment, error)
}
