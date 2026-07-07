// errorcheck

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The gofrontend used to accept this.

package p

func F2(a int32) bool {
	return a == C	// ERROR "invalid|incompatible"
}

const C = uint32(34)

// GnoError:
// line 11: 2: [function "F2" does not terminate]
// line 12: invalid operation: (mismatched types int32 and uint32)
// line 13: expected declaration, found '}'

// GoTypeCheckError:
// line 12: invalid operation: a == C (mismatched types int32 and uint32)

// GnoOverStrictError:
// line 11: 2: [function "F2" does not terminate]
// line 13: expected declaration, found '}'
