package main

// Bigdec constant arithmetic must be bounded on the denominator side.
// Repeated squaring of 1.0/3.0 doubles the denominator each step and
// would grow without limit without ratGuard.
const a0 = 1.0 / 3.0
const a1 = a0 * a0
const a2 = a1 * a1
const a3 = a2 * a2
const a4 = a3 * a3
const a5 = a4 * a4
const a6 = a5 * a5
const a7 = a6 * a6
const a8 = a7 * a7
const a9 = a8 * a8
const a10 = a9 * a9
const a11 = a10 * a10
const a12 = a11 * a11

func main() { println(a12) }

// Error:
// main/bigdec_guard_denom.gno:18:13-22: constant expression result too large: denominator exceeds 4096 bits
