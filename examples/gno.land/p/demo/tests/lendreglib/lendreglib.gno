// Package lendreglib holds library code the lend_reg filetests call.
package lendreglib

type Item struct{ N int }

// Reg is a library-owned value whose methods call a function value it holds.
type Reg struct {
	mk  func(n int) *Item
	cnt int
}

var Default = &Reg{mk: func(n int) *Item { return &Item{N: n} }}

// Make calls the library's own stored function value and returns its result.
func (r *Reg) Make(n int) *Item { return r.mk(n) }

// Each calls f once, through a library-owned receiver.
func (r *Reg) Each(f func(int)) { f(1) }

// Call calls f.
func Call(f func()) { f() }

// MakeVia calls a stored library closure through a plain function.
var mk2 = func(n int) *Item { return &Item{N: n} }

func MakeVia(n int) *Item { return mk2(n) }

// Make2 allocates directly, without a stored function value.
func (r *Reg) Make2(n int) *Item { return &Item{N: n} }
