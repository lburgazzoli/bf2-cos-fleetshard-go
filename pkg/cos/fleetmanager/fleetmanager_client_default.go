package fleetmanager

import (
	"context"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
	"net/http"
	"strconv"
)

type defaultClient struct {
	clusterId string
	api       *controlplane.APIClient
}

func (c *defaultClient) GetNamespaces(ctx context.Context, revision int64) ([]controlplane.ConnectorNamespaceDeployment, error) {
	items := make([]controlplane.ConnectorNamespaceDeployment, 0)

	for i := 1; ; i++ {
		r := c.api.ConnectorClustersAgentApi.GetClusterAsignedConnectorNamespaces(ctx, c.clusterId)
		r = r.Page(strconv.Itoa(i))
		r = r.Size(strconv.Itoa(100))
		r = r.GtVersion(revision)

		result, httpRes, err := r.Execute()
		if httpRes != nil && httpRes.Body != nil {
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

func (c *defaultClient) GetConnectors(ctx context.Context, revision int64) ([]controlplane.ConnectorDeployment, error) {
	items := make([]controlplane.ConnectorDeployment, 0)

	for i := 1; ; i++ {
		r := c.api.ConnectorClustersAgentApi.GetClusterAsignedConnectorDeployments(ctx, c.clusterId)
		r = r.Page(strconv.Itoa(i))
		r = r.Size(strconv.Itoa(100))
		r = r.GtVersion(revision)

		result, httpRes, err := r.Execute()
		if httpRes != nil && httpRes.Body != nil {
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

func (c *defaultClient) UpdateClusterStatus(ctx context.Context, status controlplane.ConnectorClusterStatus) error {
	r := c.api.ConnectorClustersAgentApi.UpdateKafkaConnectorClusterStatus(ctx, c.clusterId)
	r = r.ConnectorClusterStatus(status)

	httpRes, err := r.Execute()
	if httpRes != nil {
		_ = httpRes.Body.Close()

		if httpRes.StatusCode == http.StatusGone {
			return ResourceGone{
				error: "",
				code:  httpRes.StatusCode,
			}
		}
	}

	return err
}

func (c *defaultClient) UpdateConnectorDeploymentStatus(ctx context.Context, id string, status controlplane.ConnectorDeploymentStatus) error {
	r := c.api.ConnectorClustersAgentApi.UpdateConnectorDeploymentStatus(ctx, c.clusterId, id)
	r = r.ConnectorDeploymentStatus(status)

	httpRes, err := r.Execute()
	if httpRes != nil && httpRes.Body != nil {
		_ = httpRes.Body.Close()

		if httpRes.StatusCode == http.StatusGone {
			return ResourceGone{
				error: "",
				code:  httpRes.StatusCode,
			}
		}
	}

	return err
}
