// Package fvx backs the foreign-value probe filetests.
package fvx

import "gno.land/p/nt/avl/v0"

type T struct{ N int }

var Arr = [2]T{{N: 100}, {N: 200}}

var one = &T{N: 1}

var M = map[string]T{}

func Ptr() *T      { return one }
func New(n int) *T { return &T{N: n} }
func (t T) Val() int { return t.N }

type H [4]byte

type A [2]T

// I is an interface declared in the owning realm.
type I interface{ Val() int }

type S = struct{ N int }

func Get() T { return *one }

// Read trusts the fields of whatever *T the caller hands in.
func Read(cur realm, p *T) int { return p.N }

// E is a struct error with a value-receiver Error.
type E struct{ Msg string }

func (e E) Error() string { return e.Msg }

func Fail() error { return E{"x"} }

// Boxes holds a T by value inside an interface.
var Boxes = []any{T{N: 7}}

// Tree stores a T by value; PTree stores a *T.
var (
	Tree  avl.Tree
	PTree avl.Tree
)

func init() {
	Tree.Set("a", T{N: 8})
	PTree.Set("a", &T{N: 9})
}

// Sl is a slice of T values, for range and append probes.
var Sl = []T{{N: 1}, {N: 2}}

func (t T) Add(k int) int { return t.N + k }

// Panics panics with a T value, for recover probes.
func Panics() { panic(*one) }

// Opts takes an anonymous all-exported struct by value.
func Opts(cur realm, o struct{ N int }) int { return o.N }
func OptsNC(o struct{ N int }) int          { return o.N }

func Stats() struct{ A, B int } { return struct{ A, B int }{1, 2} }
func Pair() [2]int              { return [2]int{3, 4} }

// StatsArr returns an array of an exempt unnamed struct.
func StatsArr() [2]struct{ A, B int } {
	return [2]struct{ A, B int }{{A: 1, B: 2}, {A: 3, B: 4}}
}

var holder = struct{ H struct{ T T } }{H: struct{ T T }{T: T{N: 1}}}

// HolderField returns a pointer, so the caller copies the anonymous struct.
func HolderField() *struct{ H struct{ T T } } { return &holder }

// Hidden returns an anonymous struct with an unexported field.
func Hidden() struct{ n int } { return struct{ n int }{n: 1} }

// Do runs f inside this realm's non-crossing code.
func Do(f func()) { f() }
