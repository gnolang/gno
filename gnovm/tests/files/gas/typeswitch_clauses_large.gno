// MAXALLOC: 500000000
package main

// Type switch gas scales with clause count. Matching clause is last so every
// clause is scanned; a leading match would short-circuit and only measure
// source size.
func main() {
	var x interface{} = 1
	switch x.(type) {
	case string:
	case bool:
	case float64:
	case []int:
	case []string:
	case [2]int:
	case map[string]int:
	case int:
	}
}

// Gas:
// 5288
