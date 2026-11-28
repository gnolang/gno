package main

// A type switch clause may shadow a name declared by the type switch's init
// statement, and the type switch variable itself may shadow one.
func main() {
	var i any = 42
	switch v := "init"; y := i.(type) {
	case int:
		println("pre", v, y)
		v := y * 2
		println("post", v)
	default:
		println("def", v)
	}

	switch v := 42; v := any(v).(type) {
	case int:
		println("ts int", v)
	default:
		println("def")
	}
}

// Output:
// pre init 42
// post 84
// ts int 42
