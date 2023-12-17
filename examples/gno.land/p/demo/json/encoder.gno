package json

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// TODO: should use byte slice as input
func Flatten(m map[string]interface{}) map[string]string {
	flattened := make(map[string]string)
	var builder strings.Builder

	for k, v := range m {
		flattenValue(k, v, flattened, &builder)
	}

	return flattened
}

func flattenValue(prefix string, val interface{}, flattened map[string]string, builder *strings.Builder) {
	switch child := val.(type) {
	case map[string]interface{}:
		for k, v := range child {
			builder.Reset()
			builder.WriteString(prefix)
			builder.WriteString(".")
			builder.WriteString(k)
			flattenValue(builder.String(), v, flattened, builder)
		}
	case []interface{}:
		for i, item := range child {
			builder.Reset()
			builder.WriteString(prefix)
			builder.WriteString("[")
			builder.WriteString(strconv.Itoa(i))
			builder.WriteString("]")
			flattenValue(builder.String(), item, flattened, builder)
		}
	case string:
		flattened[prefix] = "string:\"" + child + "\""
	case int, int32, int64:
		flattened[prefix] = "number:" + fmt.Sprintf("%d", child)
	case float32, float64:
		flattened[prefix] = "number:" + fmt.Sprintf("%f", child)
	case bool:
		flattened[prefix] = "boolean:" + fmt.Sprintf("%t", child)
	default:
		flattened[prefix] = "unknown"
	}
}
