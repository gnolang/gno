// The code in this file is adapted from
// https://github.com/golang/go/blob/master/src/testing/example.go
// As such it is under the following license.
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found at the bottom of this file.

package test

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

func (opts *TestOptions) processExampleResult(name, stdout, expected string, timeSpent time.Duration, unordered, finished bool, recovered any) (passed bool) {
	passed = true
	dstr := fmtDuration(timeSpent)
	var fail string
	got := strings.TrimSpace(stdout)
	want := strings.TrimSpace(expected)
	if unordered {
		gotLines := slices.Sorted(strings.SplitSeq(got, "\n"))
		wantLines := slices.Sorted(strings.SplitSeq(want, "\n"))
		if !slices.Equal(gotLines, wantLines) && recovered == nil {
			fail = fmt.Sprintf("got:\n%s\nwant (unordered):\n%s\n", stdout, expected)
		}
	} else {
		if got != want && recovered == nil {
			fail = fmt.Sprintf("got:\n%s\nwant:\n%s\n", got, want)
		}
	}
	if fail != "" || !finished || recovered != nil {
		fmt.Fprintf(opts.Error, "--- FAIL: %s (%s)\n%s", name, dstr, fail)
		passed = false
	} else if opts.Verbose {
		fmt.Fprintf(opts.Error, "--- PASS: %s (%s)\n", name, dstr)
	}

	// XXX: We process panic recovery elsewhere
	// if recovered != nil {
	// 	   // Propagate the previously recovered result, by panicking.
	// 	   panic(recovered)
	// } else if !finished {
	// 	   panic(errNilPanicOrGoexit)
	// }

	return
}

// Copyright 2009 The Go Authors.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google LLC nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
