// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var _ = new(- /* ERROR "not a type" */ 1)
var _ = new(1 /* ERROR "not a type" */ + 1)

// GnoError:
// line 7: generic <T.(type)> requires type kind, got BigintKind
// line 8: generic <T.(type)> requires type kind, got BigintKind

// GoTypeCheckError:
// line 7
// line 8

// Clean: errorcheck
