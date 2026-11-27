// PKGPATH: gno.land/r/zrealm_wp_1
package zrealm_wp_1

import (
	"gno.land/p/demo/tests/lendlib"
	"gno.land/p/demo/tests/lendlib/keep"
)

var count int

var _ = keep.Holder{}

func run(g func()) { g() }

func runAll(fs ...func()) {
	for _, f := range fs {
		f()
	}
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

func main(cur realm) {
	// A method called on an interface holding a /p/ type from the import tree.
	var r lendlib.Runner = lendlib.Direct{}
	println("iface:", try(func() { r.Run(func() { count++ }) }), count)
	// A local helper used inside a lent literal.
	println("outer helper, lib:", try(func() { h := func() { count++ }; lendlib.Run(func() { h() }) }) != nil, count)
	println("outer helper, own:", try(func() { h := func() { count++ }; run(func() { h() }) }) != nil, count)
	// The helper called outside the literal, or its body written in it.
	println("body in literal:", try(func() { lendlib.Run(func() { count++ }) }), count)
	// A literal passed to a deferred call gets no loan.
	println("deferred:", try(func() { defer lendlib.Run(func() { count++ }) }) != nil, count)
	// A literal in a variadic slot gets no loan.
	println("variadic:", try(func() { runAll(func() { count++ }) }) != nil, count)
}

// Output:
// iface: undefined 1
// outer helper, lib: true 1
// outer helper, own: true 1
// body in literal: undefined 2
// deferred: true 2
// variadic: true 2
