// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	f[interface{}](nil)
}

func f[T any](x T) {
	var _ interface{} = x
}

// GnoOutput:

// GnoError:
// main/issue48276b.go:13:1-15:2: name T not defined in fileset with files [issue48276b.go]

// GoOutput:

// Unsupported: generics not supported in Gno
