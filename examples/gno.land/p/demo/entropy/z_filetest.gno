package main

import (
	"std"

	"gno.land/p/demo/entropy"
)

func main() {
	// initial
	println("---")
	r := entropy.New()
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())

	// should be the same
	println("---")
	r = entropy.New()
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())

	std.TestSkipHeights(1)
	println("---")
	r = entropy.New()
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())
}

// Output:
// ---
// 2206404796634629535
// 6369783909128776508
// -273967328977352263
// 3685013872379606294
// -321443877504368301
// ---
// 2206404796634629535
// 6369783909128776508
// -273967328977352263
// 3685013872379606294
// -321443877504368301
// ---
// 2389594832880495019
// 2483785090669486254
// 902493904282098385
// 5349860799650907156
// 3037689170578704503
