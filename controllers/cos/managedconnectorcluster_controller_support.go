package cos

import (
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	corev1 "k8s.io/api/core/v1"
	"strconv"
)

func computeMaxRevision[T any](items []T, annotationName string) (int64, error) {
	if len(items) == 0 {
		return 0, nil
	}

	var max int64

	for i := range items {

		var annotations map[string]string

		// TODO: there should be a better generic way of doing so
		switch item := any(items[i]).(type) {
		case corev1.Namespace:
			annotations = item.Annotations
		case cosv2.ManagedConnector:
			annotations = item.Annotations
		}

		if len(annotations) == 0 {
			continue
		}

		if r, ok := annotations[annotationName]; ok {
			rev, err := strconv.ParseInt(r, 10, 64)
			if err != nil {
				return 0, err
			}

			if rev > max {
				max = rev
			}
		}
	}

	return max, nil
}
