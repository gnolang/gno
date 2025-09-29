// Package typeutil provides utility functions for converting between different types
// and checking their states. It aims to provide consistent behavior across different
// types while remaining lightweight and dependency-free.
package typeutil

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

// stringer is the interface that wraps the String method.
type stringer interface {
	String() string
}

// ToString converts any value to its string representation.
// It supports a wide range of Go types including:
//   - Basic: string, bool
//   - Numbers: int, int8-64, uint, uint8-64, float32, float64
//   - Special: time.Time, address, []byte
//   - Slices: []T for most basic types
//   - Maps: map[string]string, map[string]any
//   - Interface: types implementing String() string
//
// Example usage:
//
//	str := typeutil.ToString(42)               // "42"
//	str = typeutil.ToString([]int{1, 2})      // "[1 2]"
//	str = typeutil.ToString(map[string]string{ // "map[a:1 b:2]"
//	    "a": "1",
//	    "b": "2",
//	})
func ToString(val any) string {
	if val == nil {
		return ""
	}

	// First check if value implements Stringer interface
	if s, ok := val.(interface{ String() string }); ok {
		return s.String()
	}

	switch v := val.(type) {
	// Pointer types - dereference and recurse
	case *string:
		if v == nil {
			return ""
		}
		return *v
	case *int:
		if v == nil {
			return ""
		}
		return strconv.Itoa(*v)
	case *bool:
		if v == nil {
			return ""
		}
		return strconv.FormatBool(*v)
	case *time.Time:
		if v == nil {
			return ""
		}
		return v.String()
	case *address:
		if v == nil {
			return ""
		}
		return string(*v)

	// String types
	case string:
		return v
	case stringer:
		return v.String()

	// Special types
	case time.Time:
		return v.String()
	case address:
		return string(v)
	case []byte:
		return string(v)
	case struct{}:
		return "{}"

	// Integer types
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)

	// Float types
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)

	// Boolean
	case bool:
		if v {
			return "true"
		}
		return "false"

	// Slice types
	case []string:
		return join(v)
	case []int:
		return join(v)
	case []int32:
		return join(v)
	case []int64:
		return join(v)
	case []float32:
		return join(v)
	case []float64:
		return join(v)
	case []any:
		return join(v)
	case []time.Time:
		return joinTimes(v)
	case []stringer:
		return join(v)
	case []address:
		return joinAddresses(v)
	case [][]byte:
		return joinBytes(v)

	// Map types with various key types
	case map[any]any, map[string]any, map[string]string, map[string]int:
		var b strings.Builder
		b.WriteString("map[")
		first := true

		switch m := v.(type) {
		case map[any]any:
			// Convert all keys to strings for consistent ordering
			keys := make([]string, 0)
			keyMap := make(map[string]any)

			for k := range m {
				keyStr := ToString(k)
				keys = append(keys, keyStr)
				keyMap[keyStr] = k
			}
			sort.Strings(keys)

			for _, keyStr := range keys {
				if !first {
					b.WriteString(" ")
				}
				origKey := keyMap[keyStr]
				b.WriteString(keyStr)
				b.WriteString(":")
				b.WriteString(ToString(m[origKey]))
				first = false
			}

		case map[string]any:
			keys := make([]string, 0)
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				if !first {
					b.WriteString(" ")
				}
				b.WriteString(k)
				b.WriteString(":")
				b.WriteString(ToString(m[k]))
				first = false
			}

		case map[string]string:
			keys := make([]string, 0)
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				if !first {
					b.WriteString(" ")
				}
				b.WriteString(k)
				b.WriteString(":")
				b.WriteString(m[k])
				first = false
			}

		case map[string]int:
			keys := make([]string, 0)
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				if !first {
					b.WriteString(" ")
				}
				b.WriteString(k)
				b.WriteString(":")
				b.WriteString(strconv.Itoa(m[k]))
				first = false
			}
		}
		b.WriteString("]")
		return b.String()

	// Default
	default:
		return "<unknown>"
	}
}

func join(slice any) string {
	if IsZero(slice) {
		return "[]"
	}

	items := ToInterfaceSlice(slice)
	if items == nil {
		return "[]"
	}

	var b strings.Builder
	b.WriteString("[")
	for i, item := range items {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(ToString(item))
	}
	b.WriteString("]")
	return b.String()
}

func joinTimes(slice []time.Time) string {
	if len(slice) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, t := range slice {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(t.String())
	}
	b.WriteString("]")
	return b.String()
}

func joinAddresses(slice []address) string {
	if len(slice) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, addr := range slice {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(string(addr))
	}
	b.WriteString("]")
	return b.String()
}

func joinBytes(slice [][]byte) string {
	if len(slice) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, bytes := range slice {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(string(bytes))
	}
	b.WriteString("]")
	return b.String()
}

// ToBool converts any value to a boolean based on common programming conventions.
// For example:
//   - Numbers: 0 is false, any other number is true
//   - Strings: "", "0", "false", "f", "no", "n", "off" are false, others are true
//   - Slices/Maps: empty is false, non-empty is true
//   - nil: always false
//   - bool: direct value
func ToBool(val any) bool {
	if IsZero(val) {
		return false
	}

	// Handle special string cases
	if str, ok := val.(string); ok {
		str = strings.ToLower(strings.TrimSpace(str))
		return str != "" && str != "0" && str != "false" && str != "f" && str != "no" && str != "n" && str != "off"
	}

	return true
}

// IsZero returns true if the value represents a "zero" or "empty" state for its type.
// For example:
//   - Numbers: 0
//   - Strings: ""
//   - Slices/Maps: empty
//   - nil: true
//   - bool: false
//   - time.Time: IsZero()
//   - address: empty string
func IsZero(val any) bool {
	if val == nil {
		return true
	}

	switch v := val.(type) {
	// Pointer types - nil pointer is zero, otherwise check pointed value
	case *bool:
		return v == nil || !*v
	case *string:
		return v == nil || *v == ""
	case *int:
		return v == nil || *v == 0
	case *time.Time:
		return v == nil || v.IsZero()
	case *address:
		return v == nil || string(*v) == ""

	// Bool
	case bool:
		return !v

	// String types
	case string:
		return v == ""
	case stringer:
		return v.String() == ""

	// Integer types
	case int:
		return v == 0
	case int8:
		return v == 0
	case int16:
		return v == 0
	case int32:
		return v == 0
	case int64:
		return v == 0
	case uint:
		return v == 0
	case uint8:
		return v == 0
	case uint16:
		return v == 0
	case uint32:
		return v == 0
	case uint64:
		return v == 0

	// Float types
	case float32:
		return v == 0
	case float64:
		return v == 0

	// Special types
	case []byte:
		return len(v) == 0
	case time.Time:
		return v.IsZero()
	case address:
		return string(v) == ""

	// Slices (check if empty)
	case []string:
		return len(v) == 0
	case []int:
		return len(v) == 0
	case []int32:
		return len(v) == 0
	case []int64:
		return len(v) == 0
	case []float32:
		return len(v) == 0
	case []float64:
		return len(v) == 0
	case []any:
		return len(v) == 0
	case []time.Time:
		return len(v) == 0
	case []address:
		return len(v) == 0
	case [][]byte:
		return len(v) == 0
	case []stringer:
		return len(v) == 0

	// Maps (check if empty)
	case map[string]string:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0

	default:
		return false // non-nil unknown types are considered non-zero
	}
}

// ToInterfaceSlice converts various slice types to []any
func ToInterfaceSlice(val any) []any {
	switch v := val.(type) {
	case []any:
		return v
	case []string:
		result := make([]any, len(v))
		for i, s := range v {
			result[i] = s
		}
		return result
	case []int:
		result := make([]any, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result
	case []int32:
		result := make([]any, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result
	case []int64:
		result := make([]any, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result
	case []float32:
		result := make([]any, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result
	case []float64:
		result := make([]any, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result
	case []bool:
		result := make([]any, len(v))
		for i, b := range v {
			result[i] = b
		}
		return result
	default:
		return nil
	}
}

// ToMapStringInterface converts a map with string keys and any value type to map[string]any
func ToMapStringInterface(m any) (map[string]any, error) {
	result := make(map[string]any)

	switch v := m.(type) {
	case map[string]any:
		return v, nil
	case map[string]string:
		for k, val := range v {
			result[k] = val
		}
	case map[string]int:
		for k, val := range v {
			result[k] = val
		}
	case map[string]int64:
		for k, val := range v {
			result[k] = val
		}
	case map[string]float64:
		for k, val := range v {
			result[k] = val
		}
	case map[string]bool:
		for k, val := range v {
			result[k] = val
		}
	case map[string][]string:
		for k, val := range v {
			result[k] = ToInterfaceSlice(val)
		}
	case map[string][]int:
		for k, val := range v {
			result[k] = ToInterfaceSlice(val)
		}
	case map[string][]any:
		for k, val := range v {
			result[k] = val
		}
	case map[string]map[string]any:
		for k, val := range v {
			result[k] = val
		}
	case map[string]map[string]string:
		for k, val := range v {
			if converted, err := ToMapStringInterface(val); err == nil {
				result[k] = converted
			} else {
				return nil, errors.New("failed to convert nested map at key: " + k)
			}
		}
	default:
		return nil, errors.New("unsupported map type: " + ToString(m))
	}

	return result, nil
}

// ToMapIntInterface converts a map with int keys and any value type to map[int]any
func ToMapIntInterface(m any) (map[int]any, error) {
	result := make(map[int]any)

	switch v := m.(type) {
	case map[int]any:
		return v, nil
	case map[int]string:
		for k, val := range v {
			result[k] = val
		}
	case map[int]int:
		for k, val := range v {
			result[k] = val
		}
	case map[int]int64:
		for k, val := range v {
			result[k] = val
		}
	case map[int]float64:
		for k, val := range v {
			result[k] = val
		}
	case map[int]bool:
		for k, val := range v {
			result[k] = val
		}
	case map[int][]string:
		for k, val := range v {
			result[k] = ToInterfaceSlice(val)
		}
	case map[int][]int:
		for k, val := range v {
			result[k] = ToInterfaceSlice(val)
		}
	case map[int][]any:
		for k, val := range v {
			result[k] = val
		}
	case map[int]map[string]any:
		for k, val := range v {
			result[k] = val
		}
	case map[int]map[int]any:
		for k, val := range v {
			result[k] = val
		}
	default:
		return nil, errors.New("unsupported map type: " + ToString(m))
	}

	return result, nil
}

// ToStringSlice converts various slice types to []string
func ToStringSlice(val any) []string {
	switch v := val.(type) {
	case []string:
		return v
	case []any:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = ToString(item)
		}
		return result
	case []int:
		result := make([]string, len(v))
		for i, n := range v {
			result[i] = strconv.Itoa(n)
		}
		return result
	case []int32:
		result := make([]string, len(v))
		for i, n := range v {
			result[i] = strconv.FormatInt(int64(n), 10)
		}
		return result
	case []int64:
		result := make([]string, len(v))
		for i, n := range v {
			result[i] = strconv.FormatInt(n, 10)
		}
		return result
	case []float32:
		result := make([]string, len(v))
		for i, n := range v {
			result[i] = strconv.FormatFloat(float64(n), 'f', -1, 32)
		}
		return result
	case []float64:
		result := make([]string, len(v))
		for i, n := range v {
			result[i] = strconv.FormatFloat(n, 'f', -1, 64)
		}
		return result
	case []bool:
		result := make([]string, len(v))
		for i, b := range v {
			result[i] = strconv.FormatBool(b)
		}
		return result
	case []time.Time:
		result := make([]string, len(v))
		for i, t := range v {
			result[i] = t.String()
		}
		return result
	case []address:
		result := make([]string, len(v))
		for i, addr := range v {
			result[i] = string(addr)
		}
		return result
	case [][]byte:
		result := make([]string, len(v))
		for i, b := range v {
			result[i] = string(b)
		}
		return result
	case []stringer:
		result := make([]string, len(v))
		for i, s := range v {
			result[i] = s.String()
		}
		return result
	case []uint:
		result := make([]string, len(v))
		for i, n := range v {
			result[i] = strconv.FormatUint(uint64(n), 10)
		}
		return result
	case []uint8:
		result := make([]string, len(v))
		for i, n := range v {
			result[i] = strconv.FormatUint(uint64(n), 10)
		}
		return result
	case []uint16:
		result := make([]string, len(v))
		for i, n := range v {
			result[i] = strconv.FormatUint(uint64(n), 10)
		}
		return result
	case []uint32:
		result := make([]string, len(v))
		for i, n := range v {
			result[i] = strconv.FormatUint(uint64(n), 10)
		}
		return result
	case []uint64:
		result := make([]string, len(v))
		for i, n := range v {
			result[i] = strconv.FormatUint(n, 10)
		}
		return result
	default:
		// Try to convert using reflection if it's a slice
		if slice := ToInterfaceSlice(val); slice != nil {
			result := make([]string, len(slice))
			for i, item := range slice {
				result[i] = ToString(item)
			}
			return result
		}
		return nil
	}
}
