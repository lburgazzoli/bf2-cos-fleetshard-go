package controller

import (
	"context"
	"fmt"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ReconciliationContext struct {
	client.Client
	types.NamespacedName

	M manager.Manager
	C context.Context

	Connector *cos.ManagedConnector
	Secret    *corev1.Secret
}

func (rc *ReconciliationContext) PatchDependant(source client.Object, target client.Object) error {
	if target.GetAnnotations() == nil {
		target.SetAnnotations(make(map[string]string))
	}

	target.GetAnnotations()[cosmeta.MetaConnectorRevision] = fmt.Sprintf("%d", rc.Connector.Spec.Deployment.ConnectorResourceVersion)
	target.GetAnnotations()[cosmeta.MetaDeploymentRevision] = fmt.Sprintf("%d", rc.Connector.Spec.Deployment.DeploymentResourceVersion)

	return resources.Apply(rc.C, rc.Client, source, target)
}

func (rc *ReconciliationContext) GetDependant(obj client.Object, opts ...client.GetOption) error {
	err := rc.Client.Get(rc.C, rc.NamespacedName, obj, opts...)
	if k8serrors.IsNotFound(err) {
		obj.SetName(rc.NamespacedName.Name)
		obj.SetNamespace(rc.NamespacedName.Namespace)

		return nil
	}

	return err
}

func (rc *ReconciliationContext) DeleteDependant(obj client.Object, opts ...client.DeleteOption) error {
	return rc.Client.Delete(rc.C, obj, opts...)
}

// Controller ---
type Controller struct {
	Owned     []client.Object
	ApplyFunc func(ReconciliationContext) error
}
