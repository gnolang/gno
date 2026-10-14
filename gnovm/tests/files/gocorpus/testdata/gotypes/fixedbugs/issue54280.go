// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

const C = 912_345_678_901_234_567_890_123_456_789_012_345_678_901_234_567_890_912_345_678_901_234_567_890_123_456_789_012_345_678_901_234_567_890_912_345_678_901_234_567_890_123_456_789_012_345_678_901_234_567_890_912 // ERROR "constant overflow"

// GnoError:
// line 7: constant overflow

// GoTypeCheckError:
// line 7

// Verdict: Clean errorcheck
