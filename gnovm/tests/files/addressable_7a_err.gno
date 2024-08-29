package main

func foo() ([]int, []string) {
	return []int{1, 2, 3}, []string{"a", "b", "c"}
}

func main() {
	_ = &foo()
}

// Error:
// main/files/addressable_7a_err.gno:8:2: getTypeOf() only supports *CallExpr with 1 result, got ([]int,[]string)
