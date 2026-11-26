package t51other

// Run is another realm's crossing function that only calls f.
func Run(cur realm, f func()) { f() }
