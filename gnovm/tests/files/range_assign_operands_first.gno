package main

type S struct{ N int }

// `for s[1], v = range s` reads each element before it writes the key, as an
// assignment evaluates its operands first: iteration 1 reads the 0 iteration 0
// wrote to s[1], before its own key writes 1 there. Go prints the same.
func main() {
	s := []int{10, 20, 30}
	var v int
	for s[1], v = range s {
		println(v)
	}
	// The map's one entry is read as 5 before its key, 0, is written to m[0].
	m := map[int]int{0: 5}
	for m[0], v = range m {
		println(v)
	}
	// A key target inside the element reads the same, whether the element is
	// copied for the value or converted, and through a pointer to an array.
	cs := []S{{10}, {20}}
	var c S
	for cs[1].N, c = range cs {
		println("copy", c.N)
	}
	us := []S{{10}, {20}}
	var u struct{ N int }
	for us[1].N, u = range us {
		println("convert", u.N)
	}
	a := [2]S{{7}, {8}}
	for a[0].N, c = range &a {
		println("arrayptr", c.N)
		break
	}
}

// Output:
// 10
// 0
// 30
// 5
// copy 10
// copy 0
// convert 10
// convert 0
// arrayptr 7
