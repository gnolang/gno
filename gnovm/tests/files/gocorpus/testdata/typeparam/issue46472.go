// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func foo[T any](d T) {
	switch v := interface{}(d).(type) {
	case string:
		if v != "x" {
			panic("unexpected v: " + v)
		}
	}

}
func main() {
	foo("x")
}

// GnoOutput:

// GnoError:
// main/issue46472.go:9:1-17:2: name T not defined in fileset with files [issue46472.go]

// GoOutput:

// Unsupported: generics not supported in Gno
