package main

func foo() ([]int, []string) {
	return []int{1, 2, 3}, []string{"a", "b", "c"}
}

func main() {
	_ = &foo()
}

// Error:
// main/addressable_7a_err.gno:8:6-12: cannot take address of multi-value call (results: ([]int,[]string))

// TypeCheckError:
// main/addressable_7a_err.gno:8:7: multiple-value foo() (value of type ([]int, []string)) in single-value context
