package blame

import (
	"fmt"
	"sort"
)

// FlattenValues takes a nested map (as parsed from YAML) and returns
// a sorted list of dot-separated key paths with their leaf values.
//
// Example:
//
//	{"image": {"repository": "nginx", "tag": "1.25"}}
//
// becomes:
//
//	[("image.repository", "nginx"), ("image.tag", "1.25")]
//
// Arrays are indexed: "hosts[0].name" for list elements.
func FlattenValues(values map[string]interface{}) []KeyValue {
	var result []KeyValue
	flattenRecursive("", values, &result)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

// KeyValue is a single flattened key-value pair.
type KeyValue struct {
	Key   string
	Value interface{}
}

func flattenRecursive(prefix string, data interface{}, result *[]KeyValue) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}
			flattenRecursive(fullKey, val, result)
		}
	case []interface{}:
		for i, item := range v {
			indexedKey := fmt.Sprintf("%s[%d]", prefix, i)
			flattenRecursive(indexedKey, item, result)
		}
	default:
		// Leaf value
		*result = append(*result, KeyValue{Key: prefix, Value: v})
	}
}
