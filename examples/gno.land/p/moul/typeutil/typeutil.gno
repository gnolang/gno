// Package typeutil provides utility functions for converting between different types
// and checking their states. It aims to provide consistent behavior across different
// types while remaining lightweight and dependency-free.
package typeutil

import (
	"fmt"
	"std"
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
//   - Special: time.Time, std.Address, []byte
//   - Slices: []T for most basic types
//   - Maps: map[string]string, map[string]interface{}
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
func ToString(val interface{}) string {
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
	case *std.Address:
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
	case std.Address:
		return string(v)
	case []byte:
		return string(v)

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
		return joinStrings(v)
	case []int:
		return joinInts(v)
	case []int32:
		return joinInt32s(v)
	case []int64:
		return joinInt64s(v)
	case []float32:
		return joinFloat32s(v)
	case []float64:
		return joinFloat64s(v)
	case []interface{}:
		return joinInterfaces(v)
	case []time.Time:
		return joinTimes(v)
	case []std.Address:
		return joinAddresses(v)
	case [][]byte:
		return joinBytes(v)
	case []stringer:
		return joinStringers(v)

	// Map types
	case map[string]interface{}, map[string]string:
		var b strings.Builder
		b.WriteString("map[")
		first := true
		switch m := v.(type) {
		case map[string]interface{}:
			for k, v := range m {
				if !first {
					b.WriteString(" ")
				}
				b.WriteString(k)
				b.WriteString(":")
				b.WriteString(ToString(v))
				first = false
			}
		case map[string]string:
			for k, v := range m {
				if !first {
					b.WriteString(" ")
				}
				b.WriteString(k)
				b.WriteString(":")
				b.WriteString(v)
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

// Helper functions for joining slices
func joinStrings(slice []string) string {
	if len(slice) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, s := range slice {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(s)
	}
	b.WriteString("]")
	return b.String()
}

func joinInts(slice []int) string {
	if len(slice) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, num := range slice {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(strconv.Itoa(num))
	}
	b.WriteString("]")
	return b.String()
}

func joinInt32s(slice []int32) string {
	if len(slice) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, num := range slice {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(strconv.FormatInt(int64(num), 10))
	}
	b.WriteString("]")
	return b.String()
}

func joinInt64s(slice []int64) string {
	if len(slice) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, num := range slice {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(strconv.FormatInt(num, 10))
	}
	b.WriteString("]")
	return b.String()
}

func joinFloat32s(slice []float32) string {
	if len(slice) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, num := range slice {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(strconv.FormatFloat(float64(num), 'f', -1, 32))
	}
	b.WriteString("]")
	return b.String()
}

func joinFloat64s(slice []float64) string {
	if len(slice) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, num := range slice {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(strconv.FormatFloat(num, 'f', -1, 64))
	}
	b.WriteString("]")
	return b.String()
}

func joinInterfaces(slice []interface{}) string {
	if len(slice) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, v := range slice {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(ToString(v))
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

func joinAddresses(slice []std.Address) string {
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

func joinStringers(slice []stringer) string {
	if len(slice) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, s := range slice {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(s.String())
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
func ToBool(val interface{}) bool {
	if val == nil {
		return false
	}

	switch v := val.(type) {
	// Pointer types - dereference and recurse
	case *bool:
		if v == nil {
			return false
		}
		return *v
	case *string:
		if v == nil {
			return false
		}
		return ToBool(*v)
	case *int:
		if v == nil {
			return false
		}
		return *v != 0
	case *time.Time:
		if v == nil {
			return false
		}
		return !v.IsZero()
	case *std.Address:
		if v == nil {
			return false
		}
		return string(*v) != ""

	// Bool
	case bool:
		return v

	// String types
	case string:
		v = strings.ToLower(strings.TrimSpace(v))
		return v != "" && v != "0" && v != "false" && v != "f" && v != "no" && v != "n" && v != "off"
	case stringer:
		return ToBool(v.String())

	// Integer types
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0

	// Float types
	case float32:
		return v != 0
	case float64:
		return v != 0

	// Special types
	case []byte:
		return len(v) > 0
	case time.Time:
		return !v.IsZero()
	case std.Address:
		return string(v) != ""

	// Slices (check if empty)
	case []string:
		return len(v) > 0
	case []int:
		return len(v) > 0
	case []int32:
		return len(v) > 0
	case []int64:
		return len(v) > 0
	case []float32:
		return len(v) > 0
	case []float64:
		return len(v) > 0
	case []interface{}:
		return len(v) > 0
	case []time.Time:
		return len(v) > 0
	case []std.Address:
		return len(v) > 0
	case [][]byte:
		return len(v) > 0
	case []stringer:
		return len(v) > 0

	// Maps (check if empty)
	case map[string]string:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0

	default:
		return true // non-nil unknown types are considered true
	}
}

// IsZero returns true if the value represents a "zero" or "empty" state for its type.
// For example:
//   - Numbers: 0
//   - Strings: ""
//   - Slices/Maps: empty
//   - nil: true
//   - bool: false
//   - time.Time: IsZero()
//   - std.Address: empty string
func IsZero(val interface{}) bool {
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
	case *std.Address:
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
	case std.Address:
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
	case []interface{}:
		return len(v) == 0
	case []time.Time:
		return len(v) == 0
	case []std.Address:
		return len(v) == 0
	case [][]byte:
		return len(v) == 0
	case []stringer:
		return len(v) == 0

	// Maps (check if empty)
	case map[string]string:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0

	default:
		return false // non-nil unknown types are considered non-zero
	}
}

// ToInterfaceSlice converts various slice types to []interface{}
func ToInterfaceSlice(val interface{}) []interface{} {
	switch v := val.(type) {
	case []interface{}:
		return v
	case []string:
		result := make([]interface{}, len(v))
		for i, s := range v {
			result[i] = s
		}
		return result
	case []int:
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result
	case []int32:
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result
	case []int64:
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result
	case []float32:
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result
	case []float64:
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result
	case []bool:
		result := make([]interface{}, len(v))
		for i, b := range v {
			result[i] = b
		}
		return result
	default:
		return nil
	}
}

// ToInterfaceMap converts a map with any value type to map[string]interface{}
func ToInterfaceMap(m interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	switch v := m.(type) {
	case map[string]interface{}:
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
	case map[string][]interface{}:
		for k, val := range v {
			result[k] = val
		}
	case map[string]map[string]interface{}:
		for k, val := range v {
			result[k] = val
		}
	case map[string]map[string]string:
		for k, val := range v {
			if converted, err := ToInterfaceMap(val); err == nil {
				result[k] = converted
			} else {
				return nil, fmt.Errorf("failed to convert nested map at key %q: %v", k, err)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported map type: %T", m)
	}

	return result, nil
}

// ToInterfaceMapInt converts a map with int keys and any value type to map[int]interface{}
func ToInterfaceMapInt(m interface{}) (map[int]interface{}, error) {
	result := make(map[int]interface{})

	switch v := m.(type) {
	case map[int]interface{}:
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
	case map[int][]interface{}:
		for k, val := range v {
			result[k] = val
		}
	case map[int]map[string]interface{}:
		for k, val := range v {
			result[k] = val
		}
	default:
		return nil, fmt.Errorf("unsupported map type: %T", m)
	}

	return result, nil
}

// ToMapStringInterface converts a map with string keys and any value type to map[string]interface{}
func ToMapStringInterface(m interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	switch v := m.(type) {
	case map[string]interface{}:
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
	case map[string][]interface{}:
		for k, val := range v {
			result[k] = val
		}
	case map[string]map[string]interface{}:
		for k, val := range v {
			result[k] = val
		}
	case map[string]map[string]string:
		for k, val := range v {
			if converted, err := ToInterfaceMap(val); err == nil {
				result[k] = converted
			} else {
				return nil, fmt.Errorf("failed to convert nested map at key %q: %v", k, err)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported map type: %T", m)
	}

	return result, nil
}

// ToMapIntInterface converts a map with int keys and any value type to map[int]interface{}
func ToMapIntInterface(m interface{}) (map[int]interface{}, error) {
	result := make(map[int]interface{})

	switch v := m.(type) {
	case map[int]interface{}:
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
	case map[int][]interface{}:
		for k, val := range v {
			result[k] = val
		}
	case map[int]map[string]interface{}:
		for k, val := range v {
			result[k] = val
		}
	case map[int]map[int]interface{}:
		for k, val := range v {
			result[k] = val
		}
	default:
		return nil, fmt.Errorf("unsupported map type: %T", m)
	}

	return result, nil
}
