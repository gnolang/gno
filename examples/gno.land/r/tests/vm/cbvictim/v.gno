package cbvictim

var Owner = "victim"

// SetOwner is exported and not crossing: callers outside must not write.
func SetOwner(s string) { Owner = s }

// Run is crossing and runs a function value the caller handed in.
func Run(cur realm, f func()) { f() }
