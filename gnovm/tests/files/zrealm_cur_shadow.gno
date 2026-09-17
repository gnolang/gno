// PKGPATH: gno.land/r/demo/curshadow
package curshadow

// A local shadowing `cur` in an inner block is realm-typed and is NOT the
// crossing parameter, so rebinding it is allowed: the write rules resolve the
// written name to its declaring block node, and a block-scoped declaration is
// not any frame's identity. Its matching half is pinned by
// zrealm_cur_shadow_call.gno: the shadow still cannot be cur-called, because
// the provenance check refuses a name whose declaring block is not a crossing
// FuncDecl/FuncLitExpr.
func main(cur realm) {
	{
		cur := cur
		cur = nil
		p := &cur
		*p = nil
		_ = cur
	}
	println("ok")
}

// Output:
// ok
