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

// extractKeys extracts all keys from a JSON object.
func extractKeys(data []byte) (keys []string, err error) {
	var (
		// stack tracks nested structures in JSON, such as objects or arrays.
		// If the top of the stack is an CurlyOpenToken ({),
		// it confirms we're inside an object.
		stack []byte
		// inString flag indicates whether the parser is currently within a string.
		inString     bool
		potentialKey string
		// afterString flag indicates whether the last parsed token was a string,
		// suggesting we may be just after a key in a key-value pair.
		afterString bool
	)

	for i := 0; i < len(data); i++ {
		switch data[i] {
		case CurlyOpenToken, SquareOpenToken:
			stack = append(stack, data[i])
		case CurlyCloseToken, SquareCloseToken:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case DoublyQuoteToken:
			if inString {
				// End of string
				inString = false
				afterString = true
			} else {
				// Start of string
				inString = true
				potentialKey = ""
				if len(stack) > 0 && stack[len(stack)-1] == CurlyOpenToken {
					afterString = false
				}
			}
		case ColonToken:
			// checks if the parser is currently within a key-value pair in a JSON object.
			if afterString && len(stack) > 0 && stack[len(stack)-1] == CurlyOpenToken {
				keys = append(keys, potentialKey)
				afterString = false
			}
		default:
			if inString {
				potentialKey += string(data[i])
			}
			if isNotWhiteSpaceOrControlChar(inString, data[i]) {
				afterString = false
			}
		}
	}

	return keys, nil
}

func isNotWhiteSpaceOrControlChar(inString bool, char byte) bool {
	return !inString && char != WhiteSpaceToken && char != NewLineToken && char != CarriageReturnToken && char != TabToken
}
