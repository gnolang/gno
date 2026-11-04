// PKGPATH: gno.land/r/demo/curshadow
package curshadow

// A local shadowing `cur` in an inner block is realm-typed and is NOT the
// crossing parameter, so the rejection below is over-broad by construction.
// That is deliberate and worth pinning: such a local can never be cur-called
// (the provenance check resolves the name's declaring block node and requires
// a crossing FuncDecl/FuncLitExpr, so a block-scoped `cur` is refused at the
// call), so refusing to rebind it costs nothing and saves the assignment rule
// from having to work out which `cur` it is looking at.

func main(cur realm) {
	{
		cur := cur.Previous()
		cur = cur.Previous()
		_ = cur
	}
}

// Error:
// gno.land/r/demo/curshadow/zrealm_cur_shadow.gno:15:3-23: cannot reassign the crossing `cur` parameter: it names the realm this frame is executing as, and that binding is fixed for the life of the frame
