package resources

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

func SetAnnotation(target *v1.ObjectMeta, key string, val string) {
	if target.Annotations == nil {
		target.Annotations = make(map[string]string)
	}

	if key != "" && val != "" {
		target.Annotations[key] = val
	}
}
