// compile

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type Value[T any] interface {
}

func use[T any](v Value[T]) {
	_, _ = v.(int)
}

func main() {
	use[int](Value[int](1))
}

// KnownIssue:
// line 12: 2: name T not defined in fileset with files [issue53762.go]
