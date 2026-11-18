// Package fvx backs the foreign-value probe filetests.
package fvx

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
