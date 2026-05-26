package mutations

import "strings"

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func flattenKeys(values map[string]any, prefix string) []string {
	keys := make([]string, 0)
	for key, value := range values {
		full := key
		if prefix != "" {
			full = prefix + "." + key
		}
		keys = append(keys, full)

		switch typed := value.(type) {
		case map[string]any:
			keys = append(keys, flattenKeys(typed, full)...)
		case []any:
			for _, item := range typed {
				nested, ok := item.(map[string]any)
				if ok {
					keys = append(keys, flattenKeys(nested, full)...)
				}
			}
		}
	}
	return keys
}
