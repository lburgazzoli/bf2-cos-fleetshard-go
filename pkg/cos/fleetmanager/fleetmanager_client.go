package fleetmanager

import (
	"context"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
)

type Client interface {
	GetNamespaces(context.Context, string, int64) ([]controlplane.ConnectorNamespaceDeployment, error)
	GetConnectors(context.Context, string, int64) ([]controlplane.ConnectorDeployment, error)
}
