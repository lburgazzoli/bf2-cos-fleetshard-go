package controller

import (
	"context"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/patch"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Patch(
	ctx context.Context,
	c client.Client,
	source client.Object,
	target client.Object,
) error {
	// TODO: server side apply
	data, err := patch.MergePatch(source, target)
	if err != nil {
		panic(err)
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
		panic(err)
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
