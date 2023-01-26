package fleetmanager

import (
	"context"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
	"strconv"
)

type defaultClient struct {
	api *controlplane.APIClient
}

func (c *defaultClient) GetNamespaces(ctx context.Context, id string, revision int64) ([]controlplane.ConnectorNamespaceDeployment, error) {
	items := make([]controlplane.ConnectorNamespaceDeployment, 0)

	for i := 1; ; i++ {
		r := c.api.ConnectorClustersAgentApi.GetClusterAsignedConnectorNamespaces(ctx, id)
		r.Page(strconv.Itoa(i))
		r.Size(strconv.Itoa(100))
		r.GtVersion(revision)

		result, httpRes, err := r.Execute()
		if httpRes != nil {
			_ = httpRes.Body.Close()
		}

		if err != nil {
			return []controlplane.ConnectorNamespaceDeployment{}, err
		}
		if len(result.Items) == 0 {
			break
		}

		items = append(items, result.Items...)
	}

	return items, nil
}

func (c *defaultClient) GetConnectors(ctx context.Context, id string, revision int64) ([]controlplane.ConnectorDeployment, error) {
	items := make([]controlplane.ConnectorDeployment, 0)

	for i := 1; ; i++ {
		r := c.api.ConnectorClustersAgentApi.GetClusterAsignedConnectorDeployments(ctx, id)
		r.Page(strconv.Itoa(i))
		r.Size(strconv.Itoa(100))
		r.GtVersion(revision)

		result, httpRes, err := r.Execute()
		if httpRes != nil {
			_ = httpRes.Body.Close()
		}

		if err != nil {
			return []controlplane.ConnectorDeployment{}, err
		}
		if len(result.Items) == 0 {
			break
		}

		items = append(items, result.Items...)
	}

	return items, nil
}
