package cos

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	cluster Cluster,
	gv int64,
) error {
	connectors, err := cluster.GetConnectors(ctx, gv)
	if err != nil {
		return errors.Wrapf(err, "failure polling for connectors")
	}

	for i := range connectors {
		r.l.Info("connector", "id", connectors[i].Id, "revision", connectors[i].Metadata.ResourceVersion)

		c := cosv2.ManagedConnector{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "mctr-" + *connectors[i].Spec.NamespaceId,
				Name:      "mctr-" + *connectors[i].Id,
			},
		}
		if err := resources.Get(ctx, r.Client, &c); err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		s := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "mctr-" + *connectors[i].Spec.NamespaceId,
				Name:      "mctr-" + *connectors[i].Id,
			},
		}
		if err := resources.Get(ctx, r.Client, &s); err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		newS := s.DeepCopy()
		newC := c.DeepCopy()
		uow := xid.New().String()

		newC.Labels = map[string]string{
			cosmeta.MetaClusterID:          cluster.Parameters.ClusterID,
			cosmeta.MetaNamespaceID:        *connectors[i].Spec.NamespaceId,
			cosmeta.MetaDeploymentID:       *connectors[i].Id,
			cosmeta.MetaDeploymentRevision: fmt.Sprintf("%d", connectors[i].Metadata.ResourceVersion),
			cosmeta.MetaConnectorID:        *connectors[i].Spec.ConnectorId,
			cosmeta.MetaConnectorRevision:  fmt.Sprintf("%d", connectors[i].Spec.ConnectorResourceVersion),
			cosmeta.MetaUnitOfWork:         uow,
		}
		newS.Labels = map[string]string{
			cosmeta.MetaClusterID:          cluster.Parameters.ClusterID,
			cosmeta.MetaNamespaceID:        *connectors[i].Spec.NamespaceId,
			cosmeta.MetaDeploymentID:       *connectors[i].Id,
			cosmeta.MetaDeploymentRevision: fmt.Sprintf("%d", connectors[i].Metadata.ResourceVersion),
			cosmeta.MetaConnectorID:        *connectors[i].Spec.ConnectorId,
			cosmeta.MetaConnectorRevision:  fmt.Sprintf("%d", connectors[i].Spec.ConnectorResourceVersion),
			cosmeta.MetaUnitOfWork:         uow,
		}
		newS.Data = make(map[string][]byte)
		newS.StringData = make(map[string]string)

		newS.Data["sa_client_id"] = []byte(connectors[i].Spec.ServiceAccount.ClientId)
		newS.Data["sa_client_secret"] = []byte(connectors[i].Spec.ServiceAccount.ClientSecret)

		for k, v := range connectors[i].Spec.ConnectorSpec {
			switch d := v.(type) {
			case map[string]interface{}:
				switch dv := d["value"].(type) {
				case string:
					connectors[i].Spec.ConnectorSpec[k] = fmt.Sprintf("{{base64:%s}}", k)
					newS.Data[k] = []byte(dv)
				case []byte:
					connectors[i].Spec.ConnectorSpec[k] = fmt.Sprintf("{{base64:%s}}", k)
					newS.Data[k] = dv
				default:
					return fmt.Errorf("unsupported value type %v", dv)
				}
			}
		}

		d, err := json.Marshal(connectors[i].Spec.ConnectorSpec)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(d, &newC.Spec.Config); err != nil {
			return err
		}

		//
		// connector
		//

		patched, err := resources.Apply(ctx, r.Client, &c, newC)
		if err != nil {
			return err
		}

		r.l.Info(
			"connector",
			"id", connectors[i].Id,
			"revision", connectors[i].Metadata.ResourceVersion,
			"patched", patched)

		//
		// Secret
		//

		patched, err = resources.Apply(ctx, r.Client, &s, newS)
		if err != nil {
			return err
		}
		if err := controllerutil.SetOwnerReference(newC, newS, r.Scheme); err != nil {
			return err
		}

		r.l.Info(
			"secret",
			"id", connectors[i].Id,
			"revision", connectors[i].Metadata.ResourceVersion,
			"patched", patched)
	}

	return nil
}
