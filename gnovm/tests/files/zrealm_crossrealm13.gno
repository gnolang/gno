package main

import (
	"chain/runtime"
	"testing"
)

func main() {
	println("From main:")
	PrintRealm()
	println(pad("CurrentRealm:"), runtime.CurrentRealm())
	println(pad("PreviousRealm:"), runtime.PreviousRealm())
	println("")

	println("From g1user origin:")
	testing.SetRealm(testing.NewUserRealm("g1user"))
	println(pad("CurrentRealm:"), runtime.CurrentRealm())
	func() {
		defer func() {
			r := recover()
			println(pad("PreviousRealm:"), r)
		}()
		println(pad("PreviousRealm:"), runtime.PreviousRealm())
	}()
	println("")

	println("From gno.land/r/sys/users realm:")
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
	PrintRealm()
	println(pad("CurrentRealm:"), runtime.CurrentRealm())
	println(pad("PreviousRealm:"), runtime.PreviousRealm())
}

func pad(s string) string {
	for len(s) < 26 {
		s += " "
	}
	return s
}

func PrintRealm() {
	println(pad("PR() CurrentRealm:"), runtime.CurrentRealm())
	println(pad("PR() PreviousRealm:"), runtime.PreviousRealm())
}

// Output:
// From main:
// PR() CurrentRealm:         CodeRealm{ g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4, main }
// PR() PreviousRealm:        UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
// CurrentRealm:              CodeRealm{ g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4, main }
// PreviousRealm:             UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
//
// From g1user origin:
// CurrentRealm:              UserRealm{ g1user }
// PreviousRealm:             frame not found: cannot seek beyond origin caller override
//
// From gno.land/r/sys/users realm:
// PR() CurrentRealm:         CodeRealm{ g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4, gno.land/r/sys/users }
// PR() PreviousRealm:        CodeRealm{ g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4, main }
// CurrentRealm:              CodeRealm{ g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4, gno.land/r/sys/users }
// PreviousRealm:             CodeRealm{ g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4, main }
