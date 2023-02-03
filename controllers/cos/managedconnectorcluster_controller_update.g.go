package cos

import (
	"context"
	"github.com/pkg/errors"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetmanager"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/pointer"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	resources := corev1.NamespaceList{}
	if err := r.List(ctx, &resources, client.MatchingLabels{cosmeta.MetaClusterID: c.Parameters.ClusterID}); err != nil {
		return err
	}

	for n := range resources.Items {
		status.Namespaces = append(status.Namespaces, fleetmanager.PresentConnectorNamespaceDeploymentStatus(resources.Items[n]))
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

	resources := cosv2.ManagedConnectorList{}
	if err := r.List(ctx, &resources, client.MatchingLabels{cosmeta.MetaClusterID: c.Parameters.ClusterID}); err != nil {
		return err
	}

	for n := range resources.Items {
		status := fleetmanager.PresentConnectorDeploymentStatus(resources.Items[n])

		if err := c.Client.UpdateConnectorDeploymentStatus(ctx, resources.Items[n].Spec.DeploymentID, status); err != nil {
			gone := fleetmanager.ResourceGone{}
			if errors.As(err, &gone) {
				r.l.Info(
					"connector gone, delete",
					"deployment_id", resources.Items[n].Spec.DeploymentID,
					"connector_id", resources.Items[n].Spec.ConnectorID)

				if err := r.Delete(ctx, res); err != nil && !k8serrors.IsNotFound(err) {
					return err
				}
			}
		}
	}

	return nil
}
