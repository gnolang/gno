package main

// A panic raised while resolving a later LHS (here: s[9] out of range) must
// leave assignments to earlier LHS in place (Go spec §Assignments:
// assignments proceed left-to-right). Complements assign_tuple_order.gno,
// where the panic comes from the assignment itself (nil *p), not resolution.

func main() {
	s := []int{0, 0}
	defer func() {
		recover()
		println(s[0])
	}()
	s[0], s[9] = 1, 2
}

// Output:
// 1
