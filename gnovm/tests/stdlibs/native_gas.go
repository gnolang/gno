package stdlibs

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// Test-stdlib gas entries. These natives only run in test mode (via
// teststdlibs.NativeResolver) and never on the live chain. Conservative
// flat values are sufficient — they exist so chargeNativeGas in gnolang
// doesn't panic when TestFiles or other test harnesses exercise them.
//
// TODO: bench through the dispatcher properly (see
// gnovm/cmd/calibrate/) and replace these flats with linear fits where
// the native has variable cost (fmt.valueOfInternal scales with input
// kind, fmt.mapKeyValues scales with map size, os.write scales with
// payload bytes, etc.). Current values are order-of-magnitude estimates
// chosen to overcharge by ~5-10× rather than undercharge.

var testNativeGas = []struct {
	pkg, fn string
	base    int64
}{
	// chain/runtime test overrides intentionally OMITTED here —
	// gnostdlibs already registered AssertOriginCall and getRealm at
	// init, and the production gas values cover the test variant fine.
	// Duplicates would panic via RegisterNativeGas.

	// fmt — typedvalue inspection helpers used by fmt.Println/Sprintf.
	// Conservative: 5000ns flat covers the worst-case kind switch.
	{"fmt", "typeString", 100},
	{"fmt", "valueOfInternal", 5000},
	{"fmt", "getAddr", 100},
	{"fmt", "getPtrElem", 100},
	{"fmt", "mapKeyValues", 5000}, // TODO: linear in map size
	{"fmt", "arrayIndex", 200},
	{"fmt", "fieldByIndex", 200},
	{"fmt", "asByteSlice", 100},

	// os — write/sleep test helpers.
	{"os", "write", 1000}, // TODO: linear in len(p)
	{"os", "sleep", 100},  // simulated — not actually sleeping in tests

	// runtime — GC / MemStats test helpers.
	{"runtime", "GC", 10000},      // flushes object cache; conservatively heavy
	{"runtime", "MemStats", 1000}, // reads alloc counters

	// testing — context / assertions / regex.
	{"testing", "getContext", 200},
	{"testing", "isRealm", 100},
	{"testing", "matchString", 1000}, // TODO: regex compile + match cost varies
	{"testing", "newRealm", 500},
	{"testing", "recoverWithStacktrace", 1000},
	{"testing", "setContext", 200},
	{"testing", "setSysParamBool", 200},
	{"testing", "setSysParamStrings", 500},
	{"testing", "testIssueCoins", 500},
	{"testing", "unixNano", 50},

	// unicode — read-only character class checks; ~100ns each.
	{"unicode", "IsGraphic", 100},
	{"unicode", "IsPrint", 100},
	{"unicode", "IsUpper", 100},
	{"unicode", "SimpleFold", 100},
}

func init() {
	for _, e := range testNativeGas {
		gno.RegisterNativeGas(e.pkg, gno.Name(e.fn), &gno.NativeGasInfo{
			Base:     e.base,
			SlopeIdx: -1,
		})
	}
}
