// PKGPATH: gno.land/r/tests/vm/chaldlendlib
package chaldlendlib

import lib "gno.land/p/demo/tests/chaldlib"

var count int

func inc()      { count++ }
func show(s string, err any) { println(s, count, err) }

func main(cur realm) {
	var box []func()
	show("01 Call (control, lent)", try(func() { lib.Call(func() { count++ }) }))
	show("02 Pass", try(func() { lib.Pass(func() { count++ }) }))
	show("03 Keep", try(func() { lib.Keep(func() { count++ }, &box) }))
	show("03b stored call", try(func() { box[0]() }))
	box = nil
	show("04 KeepCall", try(func() { lib.KeepCall(func() { count++ }, &box) }))
	box = nil
	show("05 DeferCall", try(func() { lib.DeferCall(func() { count++ }) }))
	show("06 WrapCall", try(func() { lib.WrapCall(func() { count++ }) }))
	show("07 WrapKeepCall", try(func() { lib.WrapKeepCall(func() { count++ }, &box) }))
	box = nil
	show("08 PanicIt", try(func() { lib.PanicIt(func() { count++ }, &box) }))
	box = nil
	var got func()
	show("09 PanicOut", try(func() {
		defer func() { got, _ = recover().(func()) }()
		lib.PanicOut(func() { count++ })
	}))
	lib.Call(func() {}) // reuse the frame depth with a new loan
	show("09b recovered after unwind", try(func() { got() }))
	show("10 CallFn named type", try(func() { lib.CallFn(func() { count++ }) }))
	show("11 Fn(lit).Run receiver", try(func() { lib.Fn(func() { count++ }).Run() }))
	show("12 promoted Outer.Do", try(func() { lib.Outer{}.Do(func() { count++ }) }))
	show("13 method expr Inner.Do", try(func() { lib.Inner.Do(lib.Inner{}, func() { count++ }) }))
	show("14 Outer.Keep then call", try(func() { lib.Outer{}.Keep(func() { count++ }, &box); box[0]() }))
	box = nil
	show("15 struct field Ops", try(func() { lib.CallOps(lib.Ops{F: func() { count++ }}) }))
	show("16 iface Via", try(func() { lib.Via(lib.Inner{}, func() { count++ }) }))
	show("17 Branch false", try(func() { lib.Branch(func() { count++ }, inc, false) }))
	show("18 Ret", try(func() { println(lib.Ret(func() int { count++; return count })) }))
	show("19 Recurse", try(func() { lib.Recurse(func() { count++ }, 3) }))
	show("20 nested lend", try(func() { lib.Call(func() { lib.Call(func() { count++ }) }) }))
	show("21 method value", try(func() { mv := lib.Inner{}.Do; mv(func() { count++ }) }))
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// 01 Call (control, lent) 1 undefined
// 02 Pass 2 undefined
// 03 Keep 2 undefined
// 03b stored call 2 gno.land/r/tests/vm/chaldlendlib's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldlendlib) package{})).count
// 04 KeepCall 2 gno.land/r/tests/vm/chaldlendlib's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldlendlib) package{})).count
// 05 DeferCall 3 undefined
// 06 WrapCall 4 undefined
// 07 WrapKeepCall 4 gno.land/r/tests/vm/chaldlendlib's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldlendlib) package{})).count
// 08 PanicIt 4 gno.land/r/tests/vm/chaldlendlib's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldlendlib) package{})).count
// 09 PanicOut 4 undefined
// 09b recovered after unwind 4 gno.land/r/tests/vm/chaldlendlib's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldlendlib) package{})).count
// 10 CallFn named type 5 undefined
// 11 Fn(lit).Run receiver 5 gno.land/r/tests/vm/chaldlendlib's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldlendlib) package{})).count
// 12 promoted Outer.Do 6 undefined
// 13 method expr Inner.Do 7 undefined
// 14 Outer.Keep then call 7 gno.land/r/tests/vm/chaldlendlib's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldlendlib) package{})).count
// 15 struct field Ops 7 gno.land/r/tests/vm/chaldlendlib's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldlendlib) package{})).count
// 16 iface Via 8 undefined
// 17 Branch false 8 gno.land/r/tests/vm/chaldlendlib's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldlendlib) package{})).count
// 9
// 18 Ret 9 undefined
// 19 Recurse 13 undefined
// 20 nested lend 14 undefined
// 21 method value 14 gno.land/r/tests/vm/chaldlendlib's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldlendlib) package{})).count
