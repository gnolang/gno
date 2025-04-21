package main

import (
	"std"
)

func main() {
	realm1 := std.CurrentRealm()
	println(realm1)
}

// Output:
// (struct{("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm" std.Address),("" string)} std.Realm)
