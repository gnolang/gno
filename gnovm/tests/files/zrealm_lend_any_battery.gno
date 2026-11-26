// PKGPATH: gno.land/r/zrealm_lend_any_battery
package zrealm_lend_any_battery

import lib "gno.land/p/demo/tests/lendswitchlib"

var count int

func main(cur realm) {
	// legit: lent, written in the frame
	println("defer:", try(func() { lib.DeferSwitch(func() { count++ }) }), count)
	println("conv:", try(func() { lib.ConvCallee(func() { count++ }) }), count)
	println("loop:", try(func() { lib.Loop(func() { count++ }) }), count)
	// attacks
	var r any
	println("ret:", try(func() { r = lib.RetSwitch(func() { count++ }) }), count)
	println("ret-call:", try(func() { r.(func())() }), count)
	var d any
	println("into:", try(func() { lib.Into(func() { count++ }, &d) }), count)
	println("into-call:", try(func() { d.(func())() }), count)
	println("panic:", try(func() { lib.PanicCall(func() { count++ }) }), count)
	println("two:", try(func() { lib.NewR().Two(func() {}, func() { count += 100 }) }), count)
	m := lib.NewR().Two
	println("two-mv:", try(func() { m(func() {}, func() { count += 100 }) }), count)
	println("two-me:", try(func() { (*lib.R).Two(lib.NewR(), func() {}, func() { count += 100 }) }), count)
	println("twoany:", try(func() { lib.NewR().TwoAny(func() {}, func() { count += 100 }) }), count)
	ma := lib.NewR().TwoAny
	println("twoany-mv:", try(func() { ma(func() {}, func() { count += 100 }) }), count)
	println("shadow:", try(func() { lib.Shadow(func() {}, func() { count += 100 }) }), count)
	println("count:", count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// defer: undefined 1
// conv: undefined 2
// loop: undefined 5
// ret: undefined 5
// ret-call: gno.land/r/zrealm_lend_any_battery's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_battery) package{})).count 5
// into: undefined 5
// into-call: gno.land/r/zrealm_lend_any_battery's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_battery) package{})).count 5
// panic: gno.land/r/zrealm_lend_any_battery's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_battery) package{})).count 5
// two: gno.land/r/zrealm_lend_any_battery's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_battery) package{})).count 5
// two-mv: gno.land/r/zrealm_lend_any_battery's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_battery) package{})).count 5
// two-me: gno.land/r/zrealm_lend_any_battery's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_battery) package{})).count 5
// twoany: gno.land/r/zrealm_lend_any_battery's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_battery) package{})).count 5
// twoany-mv: gno.land/r/zrealm_lend_any_battery's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_battery) package{})).count 5
// shadow: gno.land/r/zrealm_lend_any_battery's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_battery) package{})).count 5
// count: 5
