package cbattack

// Bind returns a closure that calls f(s).
func Bind(f func(string), s string) func() { return func() { f(s) } }
