// PKGPATH: gno.land/r/zrealm_wp_3
package zrealm_wp_3

import "gno.land/p/nt/avl/v0"

var tree avl.Tree

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

func main(cur realm) {
	// A method value of an imported library on the realm's own object.
	println("method value:", try(func() { mv := tree.Set; mv("a", 1) }) != nil, tree.Size())
	println("direct:", try(func() { tree.Set("b", 1) }), tree.Size())
	// A literal called where it is written, deferred.
	func() { defer func() { tree.Set("c", 1) }() }()
	println("deferred literal:", tree.Size())
}

// Output:
// method value: true 0
// direct: undefined 1
// deferred literal: 2
