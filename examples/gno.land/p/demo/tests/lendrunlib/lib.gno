package lendrunlib

// Run only calls f, so a literal written in place for f is lent to the call.
func Run(f func()) { f() }
