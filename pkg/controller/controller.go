package controller

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/meta"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/patch"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ReconcileFunc func(rc ReconciliationContext) error

func Apply(
	ctx context.Context,
	c client.Client,
	source client.Object,
	target client.Object,
) error {
	err := c.Create(ctx, target)
	if err == nil {
		return nil
	}
	if !k8serrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "error during create resource: %s/%s", target.GetNamespace(), target.GetName())
	}

	// TODO: server side apply
	data, err := patch.MergePatch(source, target)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return c.Patch(ctx, source, client.RawPatch(types.MergePatchType, data))
}

func PatchStatus(
	ctx context.Context,
	c client.Client,
	source client.Object,
	target client.Object,
) error {
	// TODO: server side apply
	data, err := patch.MergePatch(source, target)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return c.Status().Patch(ctx, source, client.RawPatch(types.MergePatchType, data))
}

func UpdateStatusCondition(conditions *[]metav1.Condition, conditionType string, consumer func(*metav1.Condition)) {
	c := meta.FindStatusCondition(*conditions, conditionType)
	if c == nil {
		c = &metav1.Condition{
			Type: conditionType,
		}
	}

	consumer(c)

	meta.SetStatusCondition(conditions, *c)

}

// FindStatusCondition finds the conditionType in conditions.
func FindStatusCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}

	return nil
}

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

	return Apply(rc.C, rc.Client, source, target)
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
