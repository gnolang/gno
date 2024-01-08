package main

import (
	"errors"
	"fmt"
	"strings"

	"gno.land/p/demo/ufmt"
)

func Marshal(data map[string]interface{}) (string, error) {
	var sb strings.Builder
	sb.WriteString("{")

	isFirst := true
	for key, value := range data {
		if !isFirst {
			sb.WriteString(", ")
		}
		sb.WriteString(ufmt.Sprintf("\"%s\": ", key))

		switch v := value.(type) {
		case string:
			sb.WriteString(ufmt.Sprintf("\"%s\"", v))
		case int, int32, int64:
			sb.WriteString(ufmt.Sprintf("%d", v))
		case float32, float64:
			sb.WriteString(fmt.Sprintf("%f", v))
		case bool:
			sb.WriteString(ufmt.Sprintf("%t", v))
		case nil:
			sb.WriteString("null")
		default:
			return "", errors.New("Marshal: unknown data type")
		}

		isFirst = false
	}

	sb.WriteString("}")
	return sb.String(), nil
}

func main() {
	zippedData := map[string]interface{}{
		"name":     "John Doe",
		"age":      30,
		"height":   1.8,
		"isMember": true,
		"address":  nil,
	}
	json, _ := Marshal(zippedData)
	fmt.Println(json)
}