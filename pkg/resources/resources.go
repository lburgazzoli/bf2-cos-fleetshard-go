package resources

import (
	"context"
	errors2 "github.com/pkg/errors"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/patch"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetAnnotation(target *metav1.ObjectMeta, key string, val string) {
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
) (bool, error) {
	err := c.Create(ctx, target)
	if err == nil {
		return false, nil
	}
	if !errors.IsAlreadyExists(err) {
		return false, errors2.Wrapf(err, "error during create resource: %s/%s", target.GetNamespace(), target.GetName())
	}

	// TODO: server side apply
	data, err := patch.MergePatch(source, target)
	if err != nil {
		return false, err
	}

	if len(data) == 0 {
		return false, nil
	}

	return true, c.Patch(ctx, source, client.RawPatch(types.MergePatchType, data))
}

func PatchStatus(
	ctx context.Context,
	c client.Client,
	source client.Object,
	target client.Object,
) (bool, error) {
	// TODO: server side apply
	data, err := patch.MergePatch(source, target)
	if err != nil {
		return false, err
	}

	if len(data) == 0 {
		return false, nil
	}

	return true, c.Status().Patch(ctx, source, client.RawPatch(types.MergePatchType, data))
}

func AsNamespacedName(obj client.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

func Get(ctx context.Context, c client.Reader, target client.Object, opts ...client.GetOption) error {
	return c.Get(ctx, AsNamespacedName(target), target, opts...)
}
