package main

import "filetests/extern/ifaceext"

// F has an unexported main-package FIELD named sec at depth 0 and the
// promoted ifaceext.sec METHOD at depth 1 via the embed. The identities are
// distinct: the field must not block F from implementing ifaceext.Sec, and
// f.sec from main must select the field (StructType lookup pins for the
// skip-gated-match rule on PR #5739).
type F struct {
	sec int
	ifaceext.SecImpl
}

func fieldsel(f F) int { return f.sec }

func main() {
	var f F
	f.sec = 42
	println(ifaceext.UseSec(f)) // promoted SecImpl.sec
	println(fieldsel(f))        // main's field
}

// Output:
// 7
// 42
