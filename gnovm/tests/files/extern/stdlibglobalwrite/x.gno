package stdlibglobalwrite

// counter is a package global. Bump writes it via a bare-identifier
// (NameExpr) assignment — the write path guarded by isExternalRealm. As a
// stdlib (dot-free path), this package may mutate its own global even when
// invoked from a /r/ realm (transient; stdlib is never persisted).
var counter int

func Bump() int {
	counter++
	return counter
}
