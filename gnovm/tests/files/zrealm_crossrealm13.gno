package main

import (
	"std"
	"testing"
)

func main() {
	println("From main:")
	PrintRealm()
	println(pad("CurrentRealm:"), std.CurrentRealm())
	println(pad("PreviousRealm:"), std.PreviousRealm())
	println("")

	println("From g1user origin:")
	testing.SetRealm(std.NewUserRealm("g1user"))
	println(pad("CurrentRealm:"), std.CurrentRealm())
	func() {
		defer func() {
			r := recover()
			println(pad("PreviousRealm:"), r)
		}()
		println(pad("PreviousRealm:"), std.PreviousRealm())
	}()
	println("")

	println("From gno.land/r/sys/users realm:")
	testing.SetRealm(std.NewCodeRealm("gno.land/r/sys/users"))
	PrintRealm()
	println(pad("CurrentRealm:"), std.CurrentRealm())
	println(pad("PreviousRealm:"), std.PreviousRealm())
}

func pad(s string) string {
	for len(s) < 26 {
		s += " "
	}
	return s
}

func PrintRealm() {
	println(pad("PR() CurrentRealm:"), std.CurrentRealm())
	println(pad("PR() PreviousRealm:"), std.PreviousRealm())
}

// Output:
// From main:
// PR() CurrentRealm:         (struct{("g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4" std.Address),("main" string)} std.Realm)
// PR() PreviousRealm:        (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
// CurrentRealm:              (struct{("g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4" std.Address),("main" string)} std.Realm)
// PreviousRealm:             (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
//
// From g1user origin:
// CurrentRealm:              (struct{("g1user" std.Address),("" string)} std.Realm)
// PreviousRealm:             frame not found: cannot seek beyond origin caller override
//
// From gno.land/r/sys/users realm:
// PR() CurrentRealm:         (struct{("g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4" std.Address),("gno.land/r/sys/users" string)} std.Realm)
// PR() PreviousRealm:        (struct{("g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4" std.Address),("main" string)} std.Realm)
// CurrentRealm:              (struct{("g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4" std.Address),("gno.land/r/sys/users" string)} std.Realm)
// PreviousRealm:             (struct{("g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4" std.Address),("main" string)} std.Realm)
