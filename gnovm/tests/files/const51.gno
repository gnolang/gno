package main

type T1 struct {
	x [2]string
}

type T2 struct {
	x *[2]string
}

func main() {
	t1 := T1{x: [2]string{"a", "b"}}
	t2 := T2{x: &[2]string{"a", "b"}}
	const c1 = len(t1.x)
	const c2 = len(t2.x)
	println(c1, c2)
}

// Output:
// 2 2
