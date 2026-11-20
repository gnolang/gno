// PKGPATH: gno.land/r/zrealm_saved_closure_capture
package zrealm_saved_closure_capture

// The realm copies its own saved table entry into a fresh closure and runs
// it. S3 refuses: the inner call reaches the stored value by name, and that
// value was made in another crossing call.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.CaptureFromTable(cross(cur))
	println("captured:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
