package cos

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

func computeMaxRevision[T any](items []T, annotationName string) (int64, error) {
	if len(items) == 0 {
		return 0, nil
	}

	var max int64

	for i := range items {
		item, ok := any(&items[i]).(client.Object)
		if !ok {
			continue
		}

		annotations := item.GetAnnotations()
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
