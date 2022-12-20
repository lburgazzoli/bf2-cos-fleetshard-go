package collections

func DeepCopyMap(m map[string]any) map[string]any {
	result := make(map[string]any)

	for k, v := range m {
		if val, ok := v.(map[string]any); ok {
			result[k] = DeepCopyMap(val)
			continue
		}
		if val, ok := v.([]any); ok {
			result[k] = DeepCopySlice(val)
			continue
		}

		result[k] = v
	}

	return result
}

func DeepCopySlice(s []any) []any {
	result := make([]any, 0)

	for _, v := range s {
		if val, ok := v.(map[string]any); ok {
			result = append(result, DeepCopyMap(val))
			continue
		}
		if val, ok := v.([]any); ok {
			result = append(result, DeepCopySlice(val))
			continue
		}

		result = append(result, v)
	}

	return result
}
