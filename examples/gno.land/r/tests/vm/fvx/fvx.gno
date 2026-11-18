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
