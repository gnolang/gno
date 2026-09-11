package main

// A switch clause is a nested scope, so it may shadow a name declared by the
// switch's init statement. Uses before the shadowing declaration still refer
// to the outer name, which keeps its own slot and its own value.
func main() {
	switch v := 1; v {
	case 1:
		println("pre", v)
		v := "s"
		println("post", v)
	}

	switch v := 1; v {
	case 1:
		var v string
		println("var", v == "")
	}

	switch v := 1; v {
	case 1:
		const v = "c"
		println("const", v)
	}

	switch v := 1; v {
	case 1:
		v := 2
		switch v := v * 10; v {
		case 20:
			v := "deep"
			println(v)
		}
		println("mid", v)
	}
}

// Output:
// pre 1
// post s
// var true
// const c
// deep
// mid 2
