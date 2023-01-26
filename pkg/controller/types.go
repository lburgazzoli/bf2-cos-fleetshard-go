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
	ConfigMap *corev1.ConfigMap
}

func (rc *ReconciliationContext) PatchDependant(source client.Object, target client.Object) (bool, error) {
	if target.GetAnnotations() == nil {
		target.SetAnnotations(make(map[string]string))
	}

	target.GetAnnotations()[cosmeta.MetaConnectorRevision] = fmt.Sprintf("%d", rc.Connector.Spec.Deployment.ConnectorResourceVersion)
	target.GetAnnotations()[cosmeta.MetaDeploymentRevision] = fmt.Sprintf("%d", rc.Connector.Spec.Deployment.DeploymentResourceVersion)

	return resources.Apply(rc.C, rc.Client, source, target)
}

func (rc *ReconciliationContext) GetDependant(obj client.Object, opts ...client.GetOption) error {
	nn := rc.NamespacedName
	if obj.GetNamespace() != "" {
		nn.Namespace = obj.GetNamespace()
	}
	if obj.GetName() != "" {
		nn.Name = obj.GetName()
	}

	err := rc.Client.Get(rc.C, nn, obj, opts...)
	if k8serrors.IsNotFound(err) {
		obj.SetName(nn.Name)
		obj.SetNamespace(nn.Namespace)

		return nil
	}

	return err
}

func (rc *ReconciliationContext) DeleteDependant(obj client.Object, opts ...client.DeleteOption) error {
	return rc.Client.Delete(rc.C, obj, opts...)
}

// Options ---
type Options struct {
	MetricsAddr                   string
	ProbeAddr                     string
	ProofAddr                     string
	EnableLeaderElection          bool
	ReleaseLeaderElectionOnCancel bool
	ID                            string
	Group                         string
	Type                          string
	Version                       string
}
