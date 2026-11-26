package main

import "strconv"

type T struct{ n int }

func (t *T) Inc() { t.n++ }

type W struct{ arr [2]int }

type Inner struct{ n int }

func (i Inner) Get() int { return i.n }

type Mid struct{ Inner }

type Outer struct{ Mid }

var pa = &[2]T{}

func mkpa() *[2]T { return pa }

func main() {
	// Indexing a pointer to an array is (*p)[i], so the element is
	// addressable even when the pointer is a call result.
	pa[0].Inc()
	mkpa()[1].Inc()
	p := &pa[1]
	p.Inc()
	pw := &[1]W{}
	_ = pw[0].arr[:]
	println(pa[0].n, pa[1].n)

	// A value method promoted through a pointer needs no addressable
	// receiver.
	m := &Mid{Inner{3}}
	println(m.Get())
	o := &Outer{Mid{Inner{4}}}
	println(o.Get())

	// A package-level var of another package is addressable.
	_ = &strconv.ErrRange
}

// Output:
// 1 2
// 3
// 4
