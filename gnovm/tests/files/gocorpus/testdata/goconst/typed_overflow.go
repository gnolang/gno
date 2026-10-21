// errorcheck

// Transcribed from Go's go/constant tests and issues #6389 #6392 #6393 #6394; expected output from go1.25.9.
// A constant expression with a typed operand must be representable in
// that type; gc rejects every line below, Gno folds most of them with
// wrapping machine arithmetic (only shifts are range checked).

package main

const a rune = '\U0010ffff'
const b = a * a // ERROR "overflows"

const c int8 = 100
const c2 = c * 2 // ERROR "overflows"

const d uint8 = 0
const d2 = d - 1 // ERROR "overflows"

const e int64 = 1 << 62
const e2 = e * 4 // ERROR "overflows"

const f uint64 = 1 << 63
const f2 = f + f // ERROR "overflows"

const g int32 = -2147483648
const g2 = -g // ERROR "overflows"

const k uint = 3
const k2 = k << 63 // ERROR "overflows"

const h int16 = 30000
const h2 = h + h // ERROR "overflows"

const i uint32 = 1 << 31
const i2 = i * 2 // ERROR "overflows"

// Untyped results beyond the destination type (#6389, #6393).
var u uint64 = 'b' - 'c' // ERROR "overflows"

var s uint

var v interface{} = ('a' << s) + (1 << 40) // ERROR "overflows"

func main() {
	println(('a' << s) + (1 << 40)) // ERROR "overflows"
}

// GnoError:
// line 29: constant overflows
// line 42: bigint overflows target kind
// line 45: bigint overflows target kind

// Verdict: KnownIssue errorcheck: #6392 uncaught markers: lines 11, 14, 17, 20, 23, 26, 32, 35 (typed constant folding wraps); line 38 is #6389 (untyped rune result not range checked against uint64)
