package chalclib

type Box struct{ N int }

func Bump(b *Box) { b.N += 1000 }

func MakeAdder(d int) func(*Box) { return func(b *Box) { b.N += d } }

// Pick returns the callback it is given (a typical option combinator).
func Pick(f func()) func() { return f }

type Registry struct{ n int }

// Reg is a library-owned object: calling its methods runs as the library.
var Reg = &Registry{}

func (r *Registry) Each(b *Box, f func(*Box)) { f(b) }
func (r *Registry) Run(f func())              { f() }
