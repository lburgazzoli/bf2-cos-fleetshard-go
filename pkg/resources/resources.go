package resources

import (
	"context"
	errors2 "github.com/pkg/errors"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/patch"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetAnnotation(target *v1.ObjectMeta, key string, val string) {
	if target.Annotations == nil {
		target.Annotations = make(map[string]string)
	}

	if key != "" && val != "" {
		target.Annotations[key] = val
	}
}

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
	if !errors.IsAlreadyExists(err) {
		return errors2.Wrapf(err, "error during create resource: %s/%s", target.GetNamespace(), target.GetName())
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
