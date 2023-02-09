package cos

import (
	"context"
	"github.com/pkg/errors"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetmanager"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/pointer"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *ManagedConnectorClusterReconciler) updateClusterStatus(
	ctx context.Context,
	res *cosv2.ManagedConnectorCluster,
) error {
	c, err := r.cluster(ctx, res)
	if err != nil {
		return errors.Wrapf(err, "unable to find cluster with id %s/%s", res.Namespace, res.Name)
	}

	status := controlplane.ConnectorClusterStatus{
		Phase:      pointer.Of(controlplane.CONNECTORCLUSTERSTATE_READY),
		Platform:   &controlplane.ConnectorClusterPlatform{Type: pointer.Of("kubernetes")},
		Namespaces: make([]controlplane.ConnectorNamespaceDeploymentStatus, 0),
		Operators:  make([]controlplane.ConnectorClusterStatusOperatorsInner, 0),
	}

	namespaces, err := r.namespaces(ctx, c)
	if err != nil {
		return err
	}

	for n := range namespaces {
		status.Namespaces = append(status.Namespaces, fleetmanager.PresentConnectorNamespaceDeploymentStatus(namespaces[n]))
	}

	return c.Client.UpdateClusterStatus(ctx, status)
}

func (r *ManagedConnectorClusterReconciler) updateConnectorsStatus(
	ctx context.Context,
	res *cosv2.ManagedConnectorCluster,
) error {
	c, err := r.cluster(ctx, res)
	if err != nil {
		return errors.Wrapf(err, "unable to find cluster with id %s/%s", res.Namespace, res.Name)
	}

	connectors, err := r.connectors(ctx, c)
	if err != nil {
		return err
	}

	for i := range connectors {
		connector := connectors[i]
		status := fleetmanager.PresentConnectorDeploymentStatus(connector)

		if err := c.Client.UpdateConnectorDeploymentStatus(ctx, connector.Spec.DeploymentID, status); err != nil {
			gone := fleetmanager.ResourceGone{}
			if errors.As(err, &gone) {
				r.l.Info(
					"connector gone, delete",
					"deployment_id", connector.Spec.DeploymentID,
					"connector_id", connector.Spec.ConnectorID)

				if err := r.Delete(ctx, &connector); err != nil && !k8serrors.IsNotFound(err) {
					return err
				}
			}
		}
	}

	return nil
}
