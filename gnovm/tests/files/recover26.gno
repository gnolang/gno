// Runtime panics raised by the VM carry a value that implements the error
// interface (as in Go), so recover().(error) works for them.
package main

func check(name string, f func()) {
	defer func() {
		r := recover()
		if e, ok := r.(error); ok {
			println(name, "=> error:", e.Error())
		} else {
			println(name, "=> NOT error:", r)
		}
	}()
	f()
}

func main() {
	check("divzero", func() { a, b := 1, 0; _ = a / b })
	check("nilcall", func() { var f func(); f() })
	check("makeslice", func() { n := -1; _ = make([]int, n) })
	check("nilderef", func() { var p *int; _ = *p })
	check("index", func() { s := []int{1, 2, 3}; i := 5; _ = s[i] })
	check("slicebounds", func() { s := []int{1, 2, 3}; h := 5; _ = s[1:h] })
	check("uncomparable", func() {
		var a, b any = []int{1}, []int{1}
		_ = a == b
	})
	check("typeassert", func() { var v any = 1; _ = v.(string) })
	check("defernil", func() { var g func(); defer g() })
}

// Output:
// divzero => error: runtime error: division by zero
// nilcall => error: runtime error: call of nil function
// makeslice => error: runtime error: makeslice: len out of range
// nilderef => error: runtime error: nil pointer dereference
// index => error: runtime error: slice index out of bounds: 5 (len=3)
// slicebounds => error: runtime error: slice bounds out of range [1:5] with capacity 3
// uncomparable => error: runtime error: comparing uncomparable type []int
// typeassert => error: int is not of type string
// defernil => error: defer called a nil function
