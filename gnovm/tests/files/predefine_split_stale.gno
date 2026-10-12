// PKGPATH: gno.land/r/splitstale
package splitstale

// a depends on its later sibling b. PredefineFileSet splits this decl and
// must rebind the index entry for b to the new split part; a stale entry
// would resolve b to the original two-name decl and report a false cycle.
var a, b = b, c
var c = 1

func main() { println(a, b, c) }

// Output:
// 1 1 1
