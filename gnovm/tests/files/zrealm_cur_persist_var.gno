// PKGPATH: gno.land/r/curpersistvar
package curpersistvar

// Persisting a captured realm in a top-level realm var must be refused at
// tx finalize time. The check fires in saveObject when the realm's backing
// *HeapItemValue is visited.
var stashed realm

func main(cur realm) {
	stashed = cur
	println("assignment ok, awaiting finalize panic")
}

// Error:
// cannot persist realm value: realm values are ephemeral and tied to a call frame
