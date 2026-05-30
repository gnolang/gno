// run

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that environment variables are accessible through
// package os.

package main

import (
	"os"
	"runtime"
)

func main() {
	ga := os.Getenv("PATH")
	if runtime.GOOS == "plan9" {
		ga = os.Getenv("path")
	}
	if ga == "" {
		print("PATH is empty\n")
		os.Exit(1)
	}
	xxx := os.Getenv("DOES_NOT_EXIST")
	if xxx != "" {
		print("$DOES_NOT_EXIST=", xxx, "\n")
		os.Exit(1)
	}
}

// TypeCheckError:
// main/env.go:18:11: undefined: os.Getenv; main/env.go:19:13: undefined: runtime.GOOS; main/env.go:20:11: undefined: os.Getenv; main/env.go:24:6: undefined: os.Exit; main/env.go:26:12: undefined: os.Getenv; main/env.go:29:6: undefined: os.Exit
