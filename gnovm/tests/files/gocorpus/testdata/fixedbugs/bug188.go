// errorcheck -d=panic

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "sort"

func main() {
	sort.Sort(nil)
	var x int
	sort(x) // ERROR "package"
}

// GnoError:
// line 14: package sort cannot only be referred to in a selector expression

// GoTypeCheckError:
// line 14: use of package sort not in selector
