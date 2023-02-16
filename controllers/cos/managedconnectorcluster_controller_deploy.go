package cos

import (
	"context"
	"encoding/base64"
	"fmt"
	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
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
	cluster *Cluster,
	gv int64,
) error {
	r.l.Info("Polling namespaces", "gv", gv)

	namespaces, err := cluster.GetNamespaces(ctx, gv)
	if err != nil {
		return errors.Wrapf(err, "failure polling for namespaces")
	}

	r.l.Info("Polling namespaces", "gv", gv, "count", len(namespaces))

	for i := range namespaces {
		r.l.Info("namespace", "id", namespaces[i].Id, "revision", namespaces[i].ResourceVersion)

		var patched bool
		var err error

		//
		// NS
		//

		ns := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mctr-" + namespaces[i].Id,
			},
		}

		if err := resources.Get(ctx, r, &ns); err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		newNs := ns.DeepCopy()
		newNs.Labels = map[string]string{
			cosmeta.MetaClusterID:   cluster.Parameters.ClusterID,
			cosmeta.MetaNamespaceID: namespaces[i].Id,
		}
		newNs.Annotations = map[string]string{
			cosmeta.MetaNamespaceRevision: fmt.Sprintf("%d", namespaces[i].ResourceVersion),
		}

		patched, err = resources.Apply(ctx, r.Client, &ns, newNs)
		if err != nil {
			return err
		}

		r.l.Info(
			"namespace",
			"id", namespaces[i].Id,
			"revision", namespaces[i].ResourceVersion,
			"patched", patched)

		//
		// IP
		//

		ipSource := camelv1.IntegrationPlatform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "camel-k",
				Namespace: cluster.MCC.Namespace,
			},
		}
		ipTarget := camelv1.IntegrationPlatform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "camel-k",
				Namespace: ns.Name,
			},
		}

		if err := resources.Get(ctx, r, &ipSource); err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
		if err := resources.Get(ctx, r, &ipTarget); err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		newIP := ipTarget.DeepCopy()
		newIP.Spec = ipSource.Spec

		patched, err = resources.Apply(ctx, r.Client, &ipTarget, newIP)
		if err != nil {
			return err
		}

		//
		// Catalog
		//

		ccSource := camelv1.CamelCatalog{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "camel-catalog-1.16.0",
				Namespace: cluster.MCC.Namespace,
			},
		}
		ccTarget := camelv1.CamelCatalog{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "camel-catalog-1.16.0",
				Namespace: ns.Name,
			},
		}

		if err := resources.Get(ctx, r, &ccSource); err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
		if err := resources.Get(ctx, r, &ccTarget); err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		newCC := ccTarget.DeepCopy()
		newCC.Spec = ccSource.Spec

		patched, err = resources.Apply(ctx, r.Client, &ccTarget, newCC)
		if err != nil {
			return err
		}

		r.l.Info(
			"integration-platform",
			"namespace", namespaces[i].Id,
			"revision", namespaces[i].ResourceVersion,
			"patched", patched)
	}

	return nil
}

func (r *ManagedConnectorClusterReconciler) deployConnectors(
	ctx context.Context,
	cluster *Cluster,
	gv int64,
) error {
	r.l.Info("Polling connectors", "gv", gv)

	connectors, err := cluster.GetConnectors(ctx, gv)
	if err != nil {
		return errors.Wrapf(err, "failure polling for connectors")
	}

	r.l.Info("Polling connectors", "gv", gv, "count", len(connectors))

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
			cosmeta.MetaClusterID:    cluster.Parameters.ClusterID,
			cosmeta.MetaNamespaceID:  *connectors[i].Spec.NamespaceId,
			cosmeta.MetaDeploymentID: *connectors[i].Id,
			cosmeta.MetaConnectorID:  *connectors[i].Spec.ConnectorId,
			cosmeta.MetaOperatorType: cosmeta.OperatorTypeCamel,
		}

		newC.Annotations = map[string]string{
			cosmeta.MetaConnectorRevision:  fmt.Sprintf("%d", *connectors[i].Spec.ConnectorResourceVersion),
			cosmeta.MetaDeploymentRevision: fmt.Sprintf("%d", connectors[i].Metadata.ResourceVersion),
			cosmeta.MetaUnitOfWork:         uow,
		}

		newC.Spec.ClusterID = cluster.Parameters.ClusterID
		newC.Spec.ConnectorID = *connectors[i].Spec.ConnectorId
		newC.Spec.ConnectorResourceVersion = *connectors[i].Spec.ConnectorResourceVersion
		newC.Spec.ConnectorTypeID = *connectors[i].Spec.ConnectorTypeId
		newC.Spec.DeploymentID = *connectors[i].Id
		newC.Spec.DeploymentResourceVersion = connectors[i].Metadata.ResourceVersion
		newC.Spec.DesiredState = cosv2.DesiredStateType(*connectors[i].Spec.DesiredState)

		if connectors[i].Spec.Kafka != nil {
			newC.Spec.Kafka = cosv2.KafkaSpec{
				URL: connectors[i].Spec.Kafka.Url,
				ID:  connectors[i].Spec.Kafka.Id,
			}
		}

		if connectors[i].Spec.SchemaRegistry != nil {
			newC.Spec.ServiceRegistry = &cosv2.ServiceRegistrySpec{
				URL: connectors[i].Spec.SchemaRegistry.Url,
				ID:  connectors[i].Spec.SchemaRegistry.Id,
			}
		}

		newS.Labels = map[string]string{
			cosmeta.MetaClusterID:    cluster.Parameters.ClusterID,
			cosmeta.MetaNamespaceID:  *connectors[i].Spec.NamespaceId,
			cosmeta.MetaDeploymentID: *connectors[i].Id,
			cosmeta.MetaConnectorID:  *connectors[i].Spec.ConnectorId,
			cosmeta.MetaOperatorType: cosmeta.OperatorTypeCamel,
		}

		newS.Annotations = map[string]string{
			cosmeta.MetaConnectorRevision:  fmt.Sprintf("%d", *connectors[i].Spec.ConnectorResourceVersion),
			cosmeta.MetaDeploymentRevision: fmt.Sprintf("%d", connectors[i].Metadata.ResourceVersion),
			cosmeta.MetaUnitOfWork:         uow,
		}

		newS.Data = make(map[string][]byte)
		newS.StringData = make(map[string]string)

		cs, err := base64.StdEncoding.DecodeString(connectors[i].Spec.ServiceAccount.ClientSecret)
		if err != nil {
			return err
		}

		newS.Data[cosmeta.ServiceAccountClientID] = []byte(connectors[i].Spec.ServiceAccount.ClientId)
		newS.Data[cosmeta.ServiceAccountClientSecret] = cs

		for k, v := range connectors[i].Spec.ConnectorSpec {
			switch d := v.(type) {
			case map[string]interface{}:
				switch dv := d["value"].(type) {
				case string:
					val, err := base64.StdEncoding.DecodeString(dv)
					if err != nil {
						return err
					}

					connectors[i].Spec.ConnectorSpec[k] = fmt.Sprintf("{{%s}}", k)
					newS.Data[k] = val
				case []byte:
					val, err := base64.StdEncoding.DecodeString(string(dv))
					if err != nil {
						return err
					}

					connectors[i].Spec.ConnectorSpec[k] = fmt.Sprintf("{{%s}}", k)
					newS.Data[k] = val
				default:
					return fmt.Errorf("unsupported value type %v", dv)
				}
			}
		}

		d, err := json.Marshal(connectors[i].Spec.ConnectorSpec)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(d, &newC.Spec.DeploymentConfig); err != nil {
			return err
		}

		m, err := json.Marshal(connectors[i].Spec.ShardMetadata)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(m, &newC.Spec.DeploymentMeta); err != nil {
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
		if err := controllerutil.SetOwnerReference(newC, newS, r.Scheme()); err != nil {
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
