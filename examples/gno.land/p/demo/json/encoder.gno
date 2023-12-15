package json

import (
	"fmt"
	"strings"
)

// Flatten takes a map and returns a new one where nested maps are replaced
// by dot-delimited keys.
func Flatten(m map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			nm := Flatten(child)
			for nk, nv := range nm {
				o[k+"."+nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o
}

func Unflatten(flatMap map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range flatMap {
		parts := strings.Split(k, ".")
		lastIdx := len(parts) - 1

		currentMap := result

		for i, part := range parts {
			if i == lastIdx {
				currentMap[part] = v
			} else {
				if _, ok := currentMap[part]; !ok {
					currentMap[part] = make(map[string]interface{})
				}

				currentMap = currentMap[part].(map[string]interface{})
			}
		}
	}

	return result
}

func MapToJSONString(m map[string]interface{}) string {
    var sb strings.Builder
    sb.WriteString("{")

    isFirst := true
    for k, v := range m {
        if !isFirst {
            sb.WriteString(", ")
        }
        sb.WriteString(fmt.Sprintf("\"%s\": %v", k, v))
        isFirst = false
    }

    sb.WriteString("}")
    return sb.String()
}