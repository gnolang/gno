package main

// More shapes of a switch or if case shadowing an init name: goto back over
// the shadow, else-if with a closure made before it, a pointer and a defer
// holding the outer name, a multi-name define, a shadowed second init name,
// and the same switch run several times in a loop.

func gotoBack() {
	switch v := 1; v {
	case 1:
		n := 0
	L:
		println("L", v)
		n++
		v := 5 + n
		println("after", v)
		if n < 2 {
			goto L
		}
	}
}

func elseIf() {
	if v := 1; v > 5 {
	} else if w := v + 1; v < 5 {
		f := func() (int, int) { return v, w }
		v := w * 10
		w := v + 1
		a, b := f()
		println(v, w, a, b)
	}
}

func pointerOuter() {
	switch v := 1; v {
	case 1:
		v = 5
		p := &v
		v := 9
		*p = 6
		println("inner", v)
		fallthrough
	case 2:
		println("outer", v)
	}
}

func deferCapture() (r int) {
	switch v := 1; v {
	case 1:
		defer func() { r = v }()
		v := 50
		_ = v
	}
	return 0
}

func multiDefine() {
	switch v := 1; v {
	case 1:
		v, w := v+1, 3
		println("inner", v, w)
		fallthrough
	case 2:
		println("outer", v)
	}
}

func twoInitNames() {
	switch a, b := 1, 2; {
	case a < b:
		b := "x"
		println(a, b)
		fallthrough
	default:
		println(a, b)
	}
}

func reuseInLoop() {
	for i := 0; i < 3; i++ {
		switch v := i; v {
		case 0:
			v := 10
			println("c0", v)
			fallthrough
		case 1:
			println("c1", v)
		default:
			v := "s"
			println("d", v)
		}
	}
}

func main() {
	gotoBack()
	elseIf()
	pointerOuter()
	println("defer", deferCapture())
	multiDefine()
	twoInitNames()
	reuseInLoop()
}

// Output:
// L 1
// after 6
// L 1
// after 7
// 20 21 1 2
// inner 9
// outer 6
// defer 1
// inner 2 3
// outer 1
// 1 x
// 1 2
// c0 10
// c1 0
// c1 1
// d s
