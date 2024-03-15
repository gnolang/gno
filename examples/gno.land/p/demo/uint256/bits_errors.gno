// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package uint256

import "errors"

var (
	overflowError = errors.New("u256: integer overflow")
	divideError   = errors.New("u256: integer divide by zero")
)