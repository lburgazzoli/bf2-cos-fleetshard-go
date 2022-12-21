package controller

import (
	"context"
	"github.com/pkg/errors"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/patch"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
