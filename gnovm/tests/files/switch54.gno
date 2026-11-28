package main

// A switch or if case may shadow a name declared by its init statement with a
// type declaration. Uses before the declaration still refer to the outer name.
func main() {
	switch v := 1; v {
	case 1:
		println("pre", v)
		type v int
		var x v = 3
		println("post", int(x))
	}

	if w := "s"; w == "" {
		println("unreachable")
	} else {
		println("pre", w)
		type w struct{ n int }
		println("post", w{4}.n)
	}

	var i any = 42
	switch v := "init"; y := i.(type) {
	case int:
		println("ts pre", v, y)
		type v uint8
		println("ts post", int(v(7)))
	}

	switch v := 1; v {
	case 1:
		type v int
		fallthrough
	case 2:
		println("fell", v)
	}
}

// Output:
// pre 1
// post 3
// pre s
// post 4
// ts pre init 42
// ts post 7
// fell 1
