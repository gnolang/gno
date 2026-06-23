// MAXALLOC: 500000000
package main

// Type switch gas scales with clause count.
//
// typeswitch_clauses_small.gno: same test with 1 clause.
func main() {
	var x interface{} = 1
	switch x.(type) {
	case int:
	case string:
	case bool:
	case float64:
	case []int:
	case []string:
	case [2]int:
	case map[string]int:
	}
}

// Gas:
// 3910
