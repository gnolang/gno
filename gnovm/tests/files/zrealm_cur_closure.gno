// PKGPATH: gno.land/r/demo/curclosure
package curclosure

// The rule resolves the written name to its declaring block node, and a
// crossing function LITERAL is one of the two declaring shapes it refuses: the
// closure's `cur` names that frame's identity exactly like a FuncDecl's does.

func main(cur realm) {
	f := func(cur realm) {
		cur = cur.Previous()
	}
	_ = f
}

// Error:
// gno.land/r/demo/curclosure/zrealm_cur_closure.gno:10:3-23: cannot reassign the crossing `cur` parameter: it names the realm this frame is executing as, and that binding is fixed for the life of the frame
