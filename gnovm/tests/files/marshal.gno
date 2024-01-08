package main

import (
	"fmt"
	"strings"
)

func Marshal(data map[string]interface{}) (string, error) {
	var sb strings.Builder
	sb.WriteString("{")

	isFirst := true
	for key, value := range data {
		if !isFirst {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("\"%s\": ", key))

		switch v := value.(type) {
		case string:
			sb.WriteString(fmt.Sprintf("\"%s\"", v))
		case int64:
			sb.WriteString(fmt.Sprintf("%d", v))
		case float64:
			sb.WriteString(fmt.Sprintf("%f", v))
		case bool:
			sb.WriteString(fmt.Sprintf("%t", v))
		default:
			return "", fmt.Errorf("unsupported type")
		}

		isFirst = false
	}

	sb.WriteString("}")
	return sb.String(), nil
}

func main() {
	zippedData := map[string]interface{}{
		"name":     "John Doe",
		"age":      30.0,
		"isMember": true,
	}
	json, _ := Marshal(zippedData)
	fmt.Println(json)
}