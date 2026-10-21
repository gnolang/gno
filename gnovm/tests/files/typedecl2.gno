package main

// A type declaration in a nested block may shadow an outer variable or type;
// uses before the declaration still refer to the outer name.
func main() {
	x := 1
	{
		println(x)
		type x struct{ n int }
		q := x{9}
		println(q.n)
	}
	println(x)

	type u struct{ n int }
	{
		type u struct{ s string }
		println(u{"y"}.s)
	}
	println(u{3}.n)
}

// Output:
// 1
// 9
// 1
// y
// 3
