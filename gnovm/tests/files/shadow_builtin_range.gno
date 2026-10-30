package main

func main() {
	for len := range []int{1} {
		println(len)
	}
}

// Output:
// 0

// TypeCheckError:
// main/shadow_builtin_range.gno:4:6: builtin identifiers cannot be shadowed: len
