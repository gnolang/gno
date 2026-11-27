// PKGPATH: gno.land/r/tests/vm/lendchainok
package lendchainok

import lib "gno.land/p/demo/tests/t77lib"

// A literal lent in place still writes when each hop passes it on, by name,
// to a parameter that is itself only called or passed on. Each line sits
// beside its one-hop twin and adds the same amount.

var count int

func one(xs []int, f func(int)) {
	for _, x := range xs {
		f(x)
	}
}

func each(xs []int, f func(int))  { one(xs, f) }
func three(xs []int, f func(int)) { each(xs, f) }

func walk(n int, f func(int)) {
	if n == 0 {
		return
	}
	f(n)
	walk(n-1, f)
}

func ping(n int, f func(int)) {
	if n == 0 {
		return
	}
	f(n)
	pong(n-1, f)
}

func pong(n int, f func(int)) { ping(n, f) }

type T struct{}

func (T) Do(xs []int, f func(int))    { one(xs, f) }
func viaMethod(xs []int, f func(int)) { T{}.Do(xs, f) }

type IntFn func(int)

func eachT(xs []int, f IntFn)    { one(xs, f) }
func conv(xs []int, f func(int)) { eachT(xs, IntFn(f)) }

func wrap(xs []int, f func(int)) { one(xs, func(x int) { f(x) }) }

type Eacher interface{ Each(xs []int, f func(int)) }
type eacher struct{}

func (eacher) Each(xs []int, f func(int))      { one(xs, f) }
func viaIface(e Eacher, xs []int, f func(int)) { e.Each(xs, f) }

func main(cur realm) {
	xs := []int{1, 2, 3}
	tree := &lib.Node{Left: &lib.Node{V: 1}, V: 2, Right: &lib.Node{V: 3}}
	r := try(func() { one(xs, func(x int) { count += x }) })
	println("c01 own, one hop", count, r)
	r = try(func() { each(xs, func(x int) { count += x }) })
	println("c02 own, two hops", count, r)
	r = try(func() { three(xs, func(x int) { count += x }) })
	println("c03 own, three hops", count, r)
	r = try(func() { lib.One(xs, func(x int) { count += x }) })
	println("c04 library, one hop", count, r)
	r = try(func() { lib.Each(xs, func(x int) { count += x }) })
	println("c05 library, two hops", count, r)
	r = try(func() { lib.Three(xs, func(x int) { count += x }) })
	println("c06 library, three hops", count, r)
	r = try(func() { walk(3, func(x int) { count += x }) })
	println("c07 own recursion", count, r)
	r = try(func() { ping(3, func(x int) { count += x }) })
	println("c08 own mutual recursion", count, r)
	r = try(func() { lib.Walk(tree, func(x int) { count += x }) })
	println("c09 library recursion", count, r)
	r = try(func() { lib.Ping(3, func(x int) { count += x }) })
	println("c10 library mutual recursion", count, r)
	r = try(func() { viaMethod(xs, func(x int) { count += x }) })
	println("c11 method hop", count, r)
	r = try(func() { lib.Inner{}.Do(xs, func(x int) { count += x }) })
	println("c12 library method hop", count, r)
	r = try(func() { conv(xs, func(x int) { count += x }) })
	println("c13 converted on the way (already so at base)", count, r)
	r = try(func() { wrap(xs, func(x int) { count += x }) })
	println("c14 wrapped in a lent literal (already so at base)", count, r)
	r = try(func() { viaIface(eacher{}, xs, func(x int) { count += x }) })
	println("c15 own interface method (already so at base)", count, r)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// c01 own, one hop 6 undefined
// c02 own, two hops 12 undefined
// c03 own, three hops 18 undefined
// c04 library, one hop 24 undefined
// c05 library, two hops 30 undefined
// c06 library, three hops 36 undefined
// c07 own recursion 42 undefined
// c08 own mutual recursion 48 undefined
// c09 library recursion 54 undefined
// c10 library mutual recursion 60 undefined
// c11 method hop 66 undefined
// c12 library method hop 72 undefined
// c13 converted on the way (already so at base) 78 undefined
// c14 wrapped in a lent literal (already so at base) 84 undefined
// c15 own interface method (already so at base) 90 undefined
