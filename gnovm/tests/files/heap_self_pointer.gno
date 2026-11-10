// A pointer stored in the slot it points at makes the value graph cyclic;
// resolving it must terminate instead of recursing until the stack overflows.
package main

type S struct{ a any }

func main() {
	// a variable pointing at itself.
	var x any
	x = &x
	println(*(x.(*any)) == x)

	// a mutual cycle across two variables.
	var a, b any
	a = &b
	b = &a
	println(*(a.(*any)) == b)

	// a struct field, and an array element, pointing at themselves.
	s := &S{}
	s.a = &s.a
	println(*(s.a.(*any)) == s.a)

	arr := &[1]any{}
	arr[0] = &arr[0]
	println(*(arr[0].(*any)) == arr[0])

	// a long acyclic chain still resolves fully.
	var head any = "end"
	for i := 0; i < 100; i++ {
		v := head
		head = &v
	}
	for {
		p, ok := head.(*any)
		if !ok {
			break
		}
		head = *p
	}
	println(head)
}

// Output:
// true
// true
// true
// true
// end
