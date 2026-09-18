// PKGPATH: gno.land/r/demo/curdeclforinit
package curdeclforinit

// A for-init `:=` is a declaration like any other. initStaticBlocks1 renames
// the binding to `cur.loopvar`, so the source spelling never reaches the
// generic declaration check; it is caught under the synthesized name instead.
// The binding is inert either way — no source name can reference it as `cur`
// once the renamer is done — but the declaration is still refused, since the
// rule is about the source binding.

func f(x int, r realm) {
	for cur := r; false; {
		_ = cur
	}
}

func main() {
	f(0, nil)
	println("ok")
}

// Error:
// gno.land/r/demo/curdeclforinit/zrealm_cur_decl_forinit.gno:12:6-14: `cur` is a contextual keyword: it may only be declared as the first parameter of a crossing function
