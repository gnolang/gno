// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that interface{M()} = *interface{M()} produces a compiler error.
// Does not compile.

package main

type Inst interface {
	Next() *Inst
}

type Regexp struct {
	code  []Inst
	start Inst
}

type Start struct {
	foo *Inst
}

func (start *Start) Next() *Inst { return nil }

func AddInst(Inst) *Inst {
	print("ok in addinst\n")
	return nil
}

func main() {
	print("call addinst\n")
	var _ Inst = AddInst(new(Start)) // ERROR "pointer to interface|incompatible type"
	print("return from  addinst\n")
	var _ *Inst = new(Start) // ERROR "pointer to interface|incompatible type"
}

// GnoError:
// line 36: main.Start does not implement main.Inst (method Next has pointer receiver)

// GoTypeCheckError:
// line 34: cannot use AddInst(new(Start)) (value of type *Inst) as Inst value in variable declaration: *Inst does not implement Inst (type *Inst is pointer to interface, not interface)
// line 36: cannot use new(Start) (value of type *Start) as *Inst value in variable declaration: *Start does not implement *Inst (type *Inst is pointer to interface, not interface)
