// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

const s float32 = 0
var _ = 0 << s

// GnoPreprocessError:
// line 8: invalid operation: invalid shift count: (const (0 float32))

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects code gc + go/types accept)
