// Package fvpass is a test fixture: library handlers that hand a caller's
// function value straight back.
package fvpass

type L struct{}

func (L) Callback(undo func()) func()  { return undo }
func (L) Use(get func() func()) func() { return get() }
func (L) PickV(fs ...func()) func()    { return fs[len(fs)-1] }

func Ident(f func()) func() { return f }
