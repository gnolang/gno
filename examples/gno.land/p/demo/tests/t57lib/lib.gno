// Package t57lib holds library code the t57 filetests call.
package t57lib

// RunAny calls f through a type switch: f is only called.
func RunAny(f any) {
	switch g := f.(type) {
	case func():
		g()
	}
}

// CallAny asserts f and calls it at once.
func CallAny(f any) { f.(func())() }

// Reg is a library-owned value whose methods call function values.
type Reg struct{ mk func() int }

var Default = &Reg{mk: func() int { return 7 }}

// Make calls the library's own stored closure.
func (r *Reg) Make() int { return r.mk() }

// Each calls f once.
func (r *Reg) Each(f func(int)) { f(1) }

// Item is library-declared data a realm may store.
type Item struct{ N int }

// Box holds library closures that allocate or write through a pointer.
type Box struct {
	mkItem func(int) *Item
	set    func(*Item, int)
	bump   func(*Box)
	n      int
}

var Lib = &Box{
	mkItem: func(n int) *Item { return &Item{N: n} },
	set:    func(it *Item, n int) { it.N = n },
	bump:   func(b *Box) { b.n++ },
}

// MakeItem calls the library's stored closure, which allocates.
func (b *Box) MakeItem(n int) *Item { return b.mkItem(n) }

// Set calls the library's stored closure, which writes through it.
func (b *Box) Set(it *Item, n int) { b.set(it, n) }

// Bump calls the library's stored closure, which writes the library's Box.
func (b *Box) Bump() int { b.bump(b); return b.n }

// Fill calls the library's stored closure on an Item it allocates.
func (b *Box) Fill(n int) int {
	it := &Item{}
	b.set(it, n)
	return it.N
}
