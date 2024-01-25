package main

func main() {
	s := 1
	f := func() {
		i := 0    // no capture for i
		var j = s // s should be captured, j not
		k := s    // s should be captured, k not
		m, n := s, 0
		println(i)
		println(j)
		println(k)
		println(m)
		println(n)
	}

	f()
}

// Output:
// 0
// 1
// 1
// 1
// 0
