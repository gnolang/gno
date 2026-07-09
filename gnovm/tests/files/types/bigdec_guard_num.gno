package main

// Integer-valued bigdec constants must be bounded on the numerator side.
// Without a numerator guard, 1e10000 (denom == 1) passes ratGuard, and
// repeated squaring at const-fold time can allocate gigabytes without gas.
const x = 1e10000

func main() { println(x) }

// Error:
// main/bigdec_guard_num.gno:6:11-18: constant expression result too large: numerator exceeds 4096 bits

// TypeCheckError:
// main/bigdec_guard_num.gno:8:23: cannot use x (untyped float constant 1e+10000) as float64 value in argument to built-in println (overflows)
