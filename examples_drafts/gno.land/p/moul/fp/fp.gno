// Package fp provides functional programming utilities for Gno, enabling
// transformations, filtering, and other operations on slices of any.
//
// Example of chaining operations:
//
//	numbers := []any{1, 2, 3, 4, 5, 6}
//
//	// Define predicates, mappers and reducers
//	isEven := func(v any) bool { return v.(int)%2 == 0 }
//	double := func(v any) any { return v.(int) * 2 }
//	sum := func(a, b any) any { return a.(int) + b.(int) }
//
//	// Chain operations: filter even numbers, double them, then sum
//	evenNums := Filter(numbers, isEven)        // [2, 4, 6]
//	doubled := Map(evenNums, double)           // [4, 8, 12]
//	result := Reduce(doubled, sum, 0)          // 24
//
//	// Alternative: group by even/odd, then get even numbers
//	byMod2 := func(v any) any { return v.(int) % 2 }
//	grouped := GroupBy(numbers, byMod2)        // {0: [2,4,6], 1: [1,3,5]}
//	evens := grouped[0]                        // [2,4,6]
package fp

// Mapper is a function type that maps an element to another element.
type Mapper func(any) any

// Predicate is a function type that evaluates a condition on an element.
type Predicate func(any) bool

// Reducer is a function type that reduces two elements to a single value.
type Reducer func(any, any) any

// Filter filters elements from the slice that satisfy the given predicate.
//
// Example:
//
//	numbers := []any{-1, 0, 1, 2}
//	isPositive := func(v any) bool { return v.(int) > 0 }
//	result := Filter(numbers, isPositive) // [1, 2]
func Filter(values []any, fn Predicate) []any {
	result := []any{}
	for _, v := range values {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Map applies a function to each element in the slice.
//
// Example:
//
//	numbers := []any{1, 2, 3}
//	toString := func(v any) any { return fmt.Sprintf("%d", v) }
//	result := Map(numbers, toString) // ["1", "2", "3"]
func Map(values []any, fn Mapper) []any {
	result := make([]any, len(values))
	for i, v := range values {
		result[i] = fn(v)
	}
	return result
}

// Reduce reduces a slice to a single value by applying a function.
//
// Example:
//
//	numbers := []any{1, 2, 3, 4}
//	sum := func(a, b any) any { return a.(int) + b.(int) }
//	result := Reduce(numbers, sum, 0) // 10
func Reduce(values []any, fn Reducer, initial any) any {
	acc := initial
	for _, v := range values {
		acc = fn(acc, v)
	}
	return acc
}

// FlatMap maps each element to a collection and flattens the results.
//
// Example:
//
//	words := []any{"hello", "world"}
//	split := func(v any) any {
//	    chars := []any{}
//	    for _, c := range v.(string) {
//	        chars = append(chars, string(c))
//	    }
//	    return chars
//	}
//	result := FlatMap(words, split) // ["h","e","l","l","o","w","o","r","l","d"]
func FlatMap(values []any, fn Mapper) []any {
	result := []any{}
	for _, v := range values {
		inner := fn(v).([]any)
		result = append(result, inner...)
	}
	return result
}

// All returns true if all elements satisfy the predicate.
//
// Example:
//
//	numbers := []any{2, 4, 6, 8}
//	isEven := func(v any) bool { return v.(int)%2 == 0 }
//	result := All(numbers, isEven) // true
func All(values []any, fn Predicate) bool {
	for _, v := range values {
		if !fn(v) {
			return false
		}
	}
	return true
}

// Any returns true if at least one element satisfies the predicate.
//
// Example:
//
//	numbers := []any{1, 3, 4, 7}
//	isEven := func(v any) bool { return v.(int)%2 == 0 }
//	result := Any(numbers, isEven) // true (4 is even)
func Any(values []any, fn Predicate) bool {
	for _, v := range values {
		if fn(v) {
			return true
		}
	}
	return false
}

// None returns true if no elements satisfy the predicate.
//
// Example:
//
//	numbers := []any{1, 3, 5, 7}
//	isEven := func(v any) bool { return v.(int)%2 == 0 }
//	result := None(numbers, isEven) // true (no even numbers)
func None(values []any, fn Predicate) bool {
	for _, v := range values {
		if fn(v) {
			return false
		}
	}
	return true
}

// Chunk splits a slice into chunks of the given size.
//
// Example:
//
//	numbers := []any{1, 2, 3, 4, 5}
//	result := Chunk(numbers, 2) // [[1,2], [3,4], [5]]
func Chunk(values []any, size int) [][]any {
	if size <= 0 {
		return nil
	}
	var chunks [][]any
	for i := 0; i < len(values); i += size {
		end := i + size
		if end > len(values) {
			end = len(values)
		}
		chunks = append(chunks, values[i:end])
	}
	return chunks
}

// Find returns the first element that satisfies the predicate and a boolean indicating if an element was found.
//
// Example:
//
//	numbers := []any{1, 2, 3, 4}
//	isEven := func(v any) bool { return v.(int)%2 == 0 }
//	result, found := Find(numbers, isEven) // 2, true
func Find(values []any, fn Predicate) (any, bool) {
	for _, v := range values {
		if fn(v) {
			return v, true
		}
	}
	return nil, false
}

// Reverse reverses the order of elements in a slice.
//
// Example:
//
//	numbers := []any{1, 2, 3}
//	result := Reverse(numbers) // [3, 2, 1]
func Reverse(values []any) []any {
	result := make([]any, len(values))
	for i, v := range values {
		result[len(values)-1-i] = v
	}
	return result
}

// Zip combines two slices into a slice of pairs. If the slices have different lengths,
// extra elements from the longer slice are ignored.
//
// Example:
//
//	a := []any{1, 2, 3}
//	b := []any{"a", "b", "c"}
//	result := Zip(a, b) // [[1,"a"], [2,"b"], [3,"c"]]
func Zip(a, b []any) [][2]any {
	length := min(len(a), len(b))
	result := make([][2]any, length)
	for i := 0; i < length; i++ {
		result[i] = [2]any{a[i], b[i]}
	}
	return result
}

// Unzip splits a slice of pairs into two separate slices.
//
// Example:
//
//	pairs := [][2]any{{1,"a"}, {2,"b"}, {3,"c"}}
//	numbers, letters := Unzip(pairs) // [1,2,3], ["a","b","c"]
func Unzip(pairs [][2]any) ([]any, []any) {
	a := make([]any, len(pairs))
	b := make([]any, len(pairs))
	for i, pair := range pairs {
		a[i] = pair[0]
		b[i] = pair[1]
	}
	return a, b
}

// GroupBy groups elements based on a key returned by a Mapper.
//
// Example:
//
//	numbers := []any{1, 2, 3, 4, 5, 6}
//	byMod3 := func(v any) any { return v.(int) % 3 }
//	result := GroupBy(numbers, byMod3) // {0: [3,6], 1: [1,4], 2: [2,5]}
func GroupBy(values []any, fn Mapper) map[any][]any {
	result := make(map[any][]any)
	for _, v := range values {
		key := fn(v)
		result[key] = append(result[key], v)
	}
	return result
}

// Flatten flattens a slice of slices into a single slice.
//
// Example:
//
//	nested := [][]any{{1,2}, {3,4}, {5}}
//	result := Flatten(nested) // [1,2,3,4,5]
func Flatten(values [][]any) []any {
	result := []any{}
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
