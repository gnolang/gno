// PKGPATH: gno.land/r/demo/nilcurcall
package nilcurcall

// Calling a nil crossing func as f(cur) must not crash the preprocessor;
// the nil call is reported at runtime, like any other nil func call.
var f func(cur realm)

func main(cur realm) {
	f(cur)
}

// Error:
// runtime error: call of nil function
