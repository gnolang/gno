package main

// The for-init DEFINE is renamed `len.loopvar` by the loop-var pass; the
// check strips the suffix and refuses the source name.

func main() {
	for len := 0; len < 1; len++ {
		println(len)
	}
}

// Output:
// 0

// TypeCheckError:
// main/shadow_builtin_forinit.gno:7:6: builtin identifiers cannot be shadowed: len
