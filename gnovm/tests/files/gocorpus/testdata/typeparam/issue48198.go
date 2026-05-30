// compile

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package p

type Foo[T any] struct {
}

func (foo Foo[T]) Get()  {
}

var(
	_ = Foo[byte]{}
	_ = Foo[[]byte]{}
	_ = Foo[map[byte]rune]{}

	_ = Foo[rune]{}
	_ = Foo[[]rune]{}
	_ = Foo[map[rune]byte]{}
)

// KnownIssue:
// line 11: 2: name T not defined in fileset with files [issue48198.go]
