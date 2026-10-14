// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

// A constraint must be an interface; it cannot
// be a type parameter, for instance.
func _[A interface{ ~int }, B A /* ERROR "cannot use a type parameter as constraint" */ ]() {}

// GoTypeCheckError:
// line 9

// GoTypesOnly:
// line 9: gc expects: cannot use a type parameter as constraint

// Verdict: Clean errorcheck
