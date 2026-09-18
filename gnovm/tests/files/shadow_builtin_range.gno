package main

func main() {
	for len := range []int{1} {
		println(len)
	}
}

// Error:
// main/shadow_builtin_range.gno:4:2-6:3: builtin identifiers cannot be shadowed: len
