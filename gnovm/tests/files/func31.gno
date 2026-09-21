// run
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Adapted from Go issue 8047 (a Go-runtime stack-copier bug, not a Gno issue).
// Here it just checks that a deferred nil func panics and is recoverable.
package main

func stackit(n int) {
	if n == 0 {
		return
	}
	stackit(n - 1)
}
func main() {
	defer func() {
		// catch & ignore panic from nil defer below
		err := recover()
		if err == nil {
			panic("defer of nil func didn't panic")
		}
		println("recovered: ", err)
	}()
	defer ((func())(nil))()
	stackit(1000)
}

// Output:
// recovered:  runtime error: defer called a nil function
