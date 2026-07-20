// MAXALLOC: 500000000
package main

// Type switch charges per non-empty clause (OpCPUSwitchClause) and per case
// (OpCPUSwitchClauseCase).
func main() {
	var x interface{} = 1
	switch x.(type) {
	case int:
	}
}

// Gas:
// 3812
