// Package fp provides functional programming utilities for Gno, enabling
// transformations, filtering, and other operations on slices of interface{}.
package fp

// Mapper is a function type that maps an element to another element.
type Mapper func(interface{}) interface{}

// Predicate is a function type that evaluates a condition on an element.
type Predicate func(interface{}) bool

// Reducer is a function type that reduces two elements to a single value.
type Reducer func(interface{}, interface{}) interface{}

// Filter filters elements from the slice that satisfy the given predicate.
func Filter(values []interface{}, fn Predicate) []interface{} {
	result := []interface{}{}
	for _, v := range values {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Map applies a function to each element in the slice.
func Map(values []interface{}, fn Mapper) []interface{} {
	result := make([]interface{}, len(values))
	for i, v := range values {
		result[i] = fn(v)
	}
	return result
}

// Reduce reduces a slice to a single value by applying a function.
func Reduce(values []interface{}, fn Reducer, initial interface{}) interface{} {
	acc := initial
	for _, v := range values {
		acc = fn(acc, v)
	}
	return acc
}

// FlatMap maps each element to a collection and flattens the results.
func FlatMap(values []interface{}, fn Mapper) []interface{} {
	result := []interface{}{}
	for _, v := range values {
		inner := fn(v).([]interface{})
		result = append(result, inner...)
	}
	return result
}

// All returns true if all elements satisfy the predicate.
func All(values []interface{}, fn Predicate) bool {
	for _, v := range values {
		if !fn(v) {
			return false
		}
	}
	return true
}

// Any returns true if at least one element satisfies the predicate.
func Any(values []interface{}, fn Predicate) bool {
	for _, v := range values {
		if fn(v) {
			return true
		}
	}
	return false
}

// None returns true if no elements satisfy the predicate.
func None(values []interface{}, fn Predicate) bool {
	for _, v := range values {
		if fn(v) {
			return false
		}
	}
	return true
}

// Chunk splits a slice into chunks of the given size.
func Chunk(values []interface{}, size int) [][]interface{} {
	if size <= 0 {
		return nil
	}
	var chunks [][]interface{}
	for i := 0; i < len(values); i += size {
		end := i + size
		if end > len(values) {
			end = len(values)
		}
		chunks = append(chunks, values[i:end])
	}
	return chunks
}

// Find returns the first element that satisfies the predicate.
func Find(values []interface{}, fn Predicate) (interface{}, bool) {
	for _, v := range values {
		if fn(v) {
			return v, true
		}
	}
	return nil, false
}

// Reverse reverses the order of elements in a slice.
func Reverse(values []interface{}) []interface{} {
	result := make([]interface{}, len(values))
	for i, v := range values {
		result[len(values)-1-i] = v
	}
	return result
}

// Zip combines two slices into a slice of pairs.
func Zip(a, b []interface{}) [][2]interface{} {
	length := min(len(a), len(b))
	result := make([][2]interface{}, length)
	for i := 0; i < length; i++ {
		result[i] = [2]interface{}{a[i], b[i]}
	}
	return result
}

// Unzip splits a slice of pairs into two slices.
func Unzip(pairs [][2]interface{}) ([]interface{}, []interface{}) {
	a := make([]interface{}, len(pairs))
	b := make([]interface{}, len(pairs))
	for i, pair := range pairs {
		a[i] = pair[0]
		b[i] = pair[1]
	}
	return a, b
}

// GroupBy groups elements based on a key returned by a Mapper.
func GroupBy(values []interface{}, fn Mapper) map[interface{}][]interface{} {
	result := make(map[interface{}][]interface{})
	for _, v := range values {
		key := fn(v)
		result[key] = append(result[key], v)
	}
	return result
}

// Flatten flattens a slice of slices into a single slice.
func Flatten(values [][]interface{}) []interface{} {
	result := []interface{}{}
	for _, v := range values {
		result = append(result, v...)
	}
	return result
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
