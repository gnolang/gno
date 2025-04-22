package main

type X struct {
	Array [8]int
	Test  bool
}

type Y [8]int

type Z struct{}

type XX struct {
	X X
}

type XB X

func main() {
	x := X{}
	z := Z{}

	println("----- zero")
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	// x manipuation

	println("----- x direct ")
	x.Array[1] = 888
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- x2 := x")
	x2 := x
	x2.Array[1] = 777
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- var x3 = x")
	var x3 = x
	x3.Array[1] = 777
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- var _, x4 = 1, x")
	var _, x4 = 1, x
	x4.Array[1] = 777
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- x.manip()")
	x.manip()
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- manip(x)")
	manip(x)
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- manip2(x)")
	manip2(x)
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- manip2(xs...) with slice 1")
	xs := []X{x}
	manip2(xs...)
	println(xs[0].Array[1])
	println(xs[0].Array[2])
	println(xs[0].Test)
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)
	xs = []X{x} // reset

	println("----- manip2(xs...) with slice 2")
	xs = []X{0: x}
	manip2(xs...)
	println(xs[0].Array[1])
	println(xs[0].Array[2])
	println(xs[0].Test)
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)
	xs = []X{x} // reset

	println("----- manip2(xs...) with array 1")
	xa := [1]X{x}
	xs = xa[:]
	manip2(xs...)
	println(xs[0].Array[1])
	println(xs[0].Array[2])
	println(xs[0].Test)
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)
	xs = []X{x} // reset

	println("----- manip2(xs...) with array 2")
	xa = [1]X{0: x}
	xs = xa[:]
	manip2(xs...)
	println(xs[0].Array[1])
	println(xs[0].Array[2])
	println(xs[0].Test)
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)
	xs = []X{x} // reset

	println("----- manip2(xs...) with struct 1")
	xx := XX{x}
	xx.X.Array[1] = 123
	println(xx.X.Array[1])
	println(x.Array[1])

	println("----- manip2(xs...) with struct 2")
	xx = XX{X: x}
	xx.X.Array[1] = 123
	println(xx.X.Array[1])
	println(x.Array[1])

	println("----- manip2(xs...) with map")
	xm := map[int]X{1: x}
	xm[1].Array[1] = 123
	println(xm[1].Array[1])
	println(x.Array[1])

	println("----- manip2(xs...) with convert")
	xb := XB(x)
	xb.Array[1] = 123
	println(xb.Array[1])
	println(x.Array[1])

	println("----- z.manip(x)")
	z.manip(x)
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- z.manip2(x)")
	z.manip2(x)
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- z.manip2(xs...)")
	z.manip2(xs...)
	println(xs[0].Array[1])
	println(xs[0].Array[2])
	println(xs[0].Test)
	xs = []X{x}

	// y manipuation

	println("----- y direct")
	y := Y{}
	y[1] = 888
	println(y[1])
	println(y[2])

	println("----- y.manip()")
	y.manip()
	println(y[1])
	println(y[2])

	// x defer manipuation

	println("----- defer x.manip()")
	func() {
		defer x.manip()
	}()
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- defer manip(x)")
	func() {
		defer manip(x)
	}()
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- defer manip2(x)")
	func() {
		defer manip2(x)
	}()
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- defer manip2(xs...)")
	func() {
		defer manip2(xs...)
	}()
	println(xs[0].Array[1])
	println(xs[0].Array[2])
	println(xs[0].Test)
	xs = []X{x}

	println("----- defer z.manip(x)")
	func() {
		defer z.manip(x)
	}()
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- defer z.manip2(x)")
	func() {
		defer z.manip2(x)
	}()
	println(x.Array[1])
	println(x.Array[2])
	println(x.Test)

	println("----- defer z.manip2(xs...)")
	func() {
		defer z.manip2(xs...)
	}()
	println(xs[0].Array[1])
	println(xs[0].Array[2])
	println(xs[0].Test)
	xs = []X{x}

	// y defer manipulation

	println("----- defer y.manip()")
	func() {
		defer y.manip()
	}()
	println(y[1])
	println(y[2])

}

func (x X) manip() {
	x.Array[2] = 999
	x.Test = true
}

func manip(x X) {
	x.Array[2] = 999
	x.Test = true
}

func manip2(xs ...X) {
	xs[0].Array[2] = 999
	xs[0].Test = true
}

func (y Y) manip() {
	y[2] = 111
}

func (_ Z) manip(x X) {
	x.Array[2] = 999
	x.Test = true
}

func (_ Z) manip2(xs ...X) {
	xs[0].Array[2] = 999
	xs[0].Test = true
}

// Output:
// ----- zero
// 0
// 0
// false
// ----- x direct 
// 888
// 0
// false
// ----- x2 := x
// 888
// 0
// false
// ----- var x3 = x
// 888
// 0
// false
// ----- var _, x4 = 1, x
// 888
// 0
// false
// ----- x.manip()
// 888
// 0
// false
// ----- manip(x)
// 888
// 0
// false
// ----- manip2(x)
// 888
// 0
// false
// ----- manip2(xs...) with slice 1
// 888
// 999
// true
// 888
// 0
// false
// ----- manip2(xs...) with slice 2
// 888
// 999
// true
// 888
// 0
// false
// ----- manip2(xs...) with array 1
// 888
// 999
// true
// 888
// 0
// false
// ----- manip2(xs...) with array 2
// 888
// 999
// true
// 888
// 0
// false
// ----- manip2(xs...) with struct 1
// 123
// 888
// ----- manip2(xs...) with struct 2
// 123
// 888
// ----- manip2(xs...) with map
// 123
// 888
// ----- manip2(xs...) with convert
// 123
// 888
// ----- z.manip(x)
// 888
// 0
// false
// ----- z.manip2(x)
// 888
// 0
// false
// ----- z.manip2(xs...)
// 888
// 999
// true
// ----- y direct
// 888
// 0
// ----- y.manip()
// 888
// 0
// ----- defer x.manip()
// 888
// 0
// false
// ----- defer manip(x)
// 888
// 0
// false
// ----- defer manip2(x)
// 888
// 0
// false
// ----- defer manip2(xs...)
// 888
// 999
// true
// ----- defer z.manip(x)
// 888
// 0
// false
// ----- defer z.manip2(x)
// 888
// 0
// false
// ----- defer z.manip2(xs...)
// 888
// 999
// true
// ----- defer y.manip()
// 888
// 0

// TypeCheckError:
// main/files/issue_1096.gno:134:2: cannot assign to xm[1].Array[1] (neither addressable nor a map index expression)
