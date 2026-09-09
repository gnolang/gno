// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

// Clone returns a fresh copy of s.
// It guarantees to make a copy of s into a new allocation,
// which can be important when retaining only a small substring
// of a much larger string. Using Clone can help such programs
// use less memory. Of course, since using Clone makes a copy,
// overuse of Clone can make programs use more memory.
// Clone should typically be used only rarely, and only when
// profiling indicates that it is needed.
// For strings of length zero the string "" will be returned
// and no allocation is made.
func Clone(s string) string {
	if len(s) == 0 {
		return ""
	}
	// Custom code: Go uses unsafe.String over a freshly allocated byte
	// slice to avoid the second copy. Gno has no unsafe, so the
	// []byte -> string conversion performs that copy instead; the
	// observable behaviour (a string not aliasing s) is the same.
	return string([]byte(s))
}
