// PKGPATH: gno.land/r/tests/vm/lendchainrefused
package lendchainrefused

import (
	lib "gno.land/p/demo/tests/t77lib"
	"gno.land/r/tests/vm/t77other"
)

// Each first hop only passes f on; the shape after it must end f's loan, so
// the call that follows is refused and count does not move.

var count int
var saved func(int)

func each2(xs []int, f func(int)) {
	for _, x := range xs {
		f(x)
	}
}

func call1(f func(int)) { f(1) }

// r01: the chain ends in a store.
func store1(xs []int, f func(int)) { store2(f); saved(1) }
func store2(f func(int))           { saved = f }

// r02: the chain ends in a return.
func ret1(xs []int, f func(int)) { g := ret2(f); g(1) }
func ret2(f func(int)) func(int) { return f }

// r03: the next hop defers a call with f.
func defer1(xs []int, f func(int)) { defer2(f) }
func defer2(f func(int))           { defer call1(f) }

// r04: the next hop is a local reassigned, so reached by value.
func value1(xs []int, f func(int)) {
	g := each2
	g = each2
	g(xs, f)
}

// r05: the next hop is a package variable.
var eachVar = each2

func pkgvar1(xs []int, f func(int)) { eachVar(xs, f) }

// r06: the next hop is an interface method of another realm's type.
type Eacher interface{ Each(xs []int, f func(int)) }

func iface1(e Eacher, xs []int, f func(int)) { e.Each(xs, f) }

// r07: the next hop is another realm's plain function.
func other1(xs []int, f func(int)) { t77other.Each(xs, f) }

// r08: f is wrapped in a literal that is stored on the way.
func wrap1(xs []int, f func(int)) { store2(func(x int) { f(x) }); saved(1) }

// r09: f is converted, and the next hop stores it.
type IntFn func(int)

var savedT IntFn

func eachT(xs []int, f IntFn)     { savedT = f }
func conv1(xs []int, f func(int)) { eachT(xs, IntFn(f)); savedT(1) }

// r10: the last hop stores f on one branch only.
func branch1(xs []int, f func(int)) { branch2(f, false) }
func branch2(f func(int), keep bool) {
	if keep {
		saved = f
		return
	}
	f(1)
}

// r11: the library's chain ends in a store.
func lib1(xs []int, f func(int)) { h := &lib.Holder{}; lib.Stash(h, f); h.F(1) }

// r12: the chain stores f and returns; the copy is called afterwards.
func keep1(xs []int, f func(int)) { store2(f) }

func main(cur realm) {
	xs := []int{1, 2, 3}
	r := try(func() { each2(xs, func(x int) { count += x }) })
	println("r00 control: one hop", count, r != nil)
	r = try(func() { store1(xs, func(x int) { count += x }) })
	println("r01 store", count, r != nil)
	r = try(func() { ret1(xs, func(x int) { count += x }) })
	println("r02 return", count, r != nil)
	r = try(func() { defer1(xs, func(x int) { count += x }) })
	println("r03 deferred", count, r != nil)
	r = try(func() { value1(xs, func(x int) { count += x }) })
	println("r04 value callee", count, r != nil)
	r = try(func() { pkgvar1(xs, func(x int) { count += x }) })
	println("r05 package variable callee", count, r != nil)
	r = try(func() { iface1(t77other.NewE().(Eacher), xs, func(x int) { count += x }) })
	println("r06 interface method of another realm", count, r != nil)
	r = try(func() { other1(xs, func(x int) { count += x }) })
	println("r07 another realm", count, r != nil)
	r = try(func() { wrap1(xs, func(x int) { count += x }) })
	println("r08 wrapped, wrapper stored", count, r != nil)
	r = try(func() { conv1(xs, func(x int) { count += x }) })
	println("r09 converted, then stored", count, r != nil)
	r = try(func() { branch1(xs, func(x int) { count += x }) })
	println("r10 stores on one branch", count, r != nil)
	r = try(func() { lib1(xs, func(x int) { count += x }) })
	println("r11 library chain stores", count, r != nil)
	r = try(func() { keep1(xs, func(x int) { count += x }) })
	println("r12a store, then return", count, r != nil)
	r = try(func() { saved(1) })
	println("r12b kept copy called after return", count, r != nil)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// r00 control: one hop 6 false
// r01 store 6 true
// r02 return 6 true
// r03 deferred 6 true
// r04 value callee 6 true
// r05 package variable callee 6 true
// r06 interface method of another realm 6 true
// r07 another realm 6 true
// r08 wrapped, wrapper stored 6 true
// r09 converted, then stored 6 true
// r10 stores on one branch 6 true
// r11 library chain stores 6 true
// r12a store, then return 6 false
// r12b kept copy called after return 6 true
