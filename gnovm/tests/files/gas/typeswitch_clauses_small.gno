// MAXALLOC: 500000000
package main

// Type switch charges gas per clause via OpCPUSlopeTypeSwitchCase.
//
// typeswitch_clauses_large.gno: same test with 8 concrete clauses.
func main() {
	var x interface{} = 1
	switch x.(type) {
	case int:
	}
}

// Gas:
// 4061
