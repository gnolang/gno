// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f[G any]() int {
	return func() int {
		return func() int {
			return 0
		}()
	}()
}

func main() {
	f[int]()
}

// GnoOutput:

// GnoError:
// main/issue47684.go:18:2-8: unexpected index base kind for type func() int

// GoOutput:

// Unsupported: generics not supported in Gno
