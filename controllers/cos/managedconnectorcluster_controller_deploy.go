package cos

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *ManagedConnectorClusterReconciler) deployNamespaces(
	ctx context.Context,
	c Cluster,
	gv int64,
) error {
	namespaces, err := c.GetNamespaces(ctx, gv)
	if err != nil {
		return errors.Wrapf(err, "failure polling for namespaces")
	}

	for i := range namespaces {
		r.l.Info("namespace", "id", namespaces[i].Id, "revision", namespaces[i].ResourceVersion)

		ns := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mctr-" + namespaces[i].Id,
				Labels: map[string]string{
					cosmeta.MetaClusterID:         c.Parameters.ClusterID,
					cosmeta.MetaNamespaceID:       namespaces[i].Id,
					cosmeta.MetaNamespaceRevision: fmt.Sprintf("%d", namespaces[i].ResourceVersion),
				},
			},
		}

		newNs := ns.DeepCopy()

		patched, err := resources.Apply(ctx, r.Client, &ns, newNs)
		if err != nil {
			return err
		}

		r.l.Info(
			"namespace",
			"id", namespaces[i].Id,
			"revision", namespaces[i].ResourceVersion,
			"patched", patched)
	}

	return nil
}

func (r *ManagedConnectorClusterReconciler) deployConnectors(
	ctx context.Context,
	c Cluster,
	gv int64,
) error {
	connectors, err := c.GetConnectors(ctx, gv)
	if err != nil {
		return errors.Wrapf(err, "failure polling for connectors")
	}

	for i := range connectors {
		r.l.Info("connector", "id", connectors[i].Id, "revision", connectors[i].Metadata.ResourceVersion)

		c := cosv2.ManagedConnector{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "mctr-" + *connectors[i].Spec.NamespaceId,
				Name:      "mctr-" + *connectors[i].Id,
				Labels: map[string]string{
					cosmeta.MetaClusterID:          c.Parameters.ClusterID,
					cosmeta.MetaNamespaceID:        *connectors[i].Spec.NamespaceId,
					cosmeta.MetaDeploymentID:       *connectors[i].Id,
					cosmeta.MetaDeploymentRevision: fmt.Sprintf("%d", connectors[i].Metadata.ResourceVersion),
					cosmeta.MetaConnectorID:        *connectors[i].Spec.ConnectorId,
					cosmeta.MetaConnectorRevision:  fmt.Sprintf("%d", connectors[i].Spec.ConnectorResourceVersion),
				},
			},
		}

		newC := c.DeepCopy()

		patched, err := resources.Apply(ctx, r.Client, &c, newC)
		if err != nil {
			return err
		}

		r.l.Info(
			"connector",
			"id", connectors[i].Id,
			"revision", connectors[i].Metadata.ResourceVersion,
			"patched", patched)
	}

	return nil
}
