// Verifies the errorcheck-style path: a .go file carrying inline
// `// ERROR "regex"` markers passes when Gno rejects it with wording
// that matches at least one marker. Uses `package p` (non-main) to
// also exercise the PKGPATH+synthetic-main rescue.

package p

var x = undeclared // ERROR "undeclared|undefined|not declared"

// GnoError:
// line 8: name undeclared not defined in fileset with files [canary.go]
