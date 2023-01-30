package cos

import (
	"context"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetmanager"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/pointer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ManagedConnectorClusterReconciler) updateClusterStatus(
	ctx context.Context,
	named types.NamespacedName,
	mcc *cosv2.ManagedConnectorCluster,
) error {
	c, err := r.cluster(ctx, named, mcc)
	if err != nil {
		return err
	}

	status := controlplane.ConnectorClusterStatus{
		Phase:      pointer.Of(controlplane.CONNECTORCLUSTERSTATE_READY),
		Platform:   &controlplane.ConnectorClusterPlatform{Type: pointer.Of("kubernetes")},
		Namespaces: make([]controlplane.ConnectorNamespaceDeploymentStatus, 0),
		Operators:  make([]controlplane.ConnectorClusterStatusOperatorsInner, 0),
	}

	namespaces := corev1.NamespaceList{}
	if err := r.List(ctx, &namespaces, client.MatchingLabels{cosmeta.MetaClusterID: c.Parameters.ClusterID}); err != nil {
		return err
	}

	for n := range namespaces.Items {
		status.Namespaces = append(status.Namespaces, fleetmanager.PresentConnectorNamespaceDeploymentStatus(namespaces.Items[n]))
	}

	return c.Client.UpdateClusterStatus(ctx, status)
}
