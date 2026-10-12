package main

type T1 struct {
	x [2]string
}

type T2 struct {
	x *[2]string
}

func main() {
	t1 := T1{x: [2]string{"a", "b"}}
	_ = t1 // keep the variable used; the test is about something else
	t2 := T2{x: &[2]string{"a", "b"}}
	const c1 = len(t1.x)
	const c2 = len(t2.x)
	println(c1, c2)
}

// Error:
// main/const51.gno:14:2-4: declared and not used: t2
