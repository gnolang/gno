// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// stand-alone test to ensure case is triggered

package issue26390

type A = T

func (t *T) m() *A { return t }

type T struct{}

// GnoPreprocessError:
// line 9: StaticBlock.Define2(A) cannot change .V

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects code gc + go/types accept)
