package main

import "math"

func main() {
    const f = math.MaxFloat64
    println(float32(f))
}

// Error:
// main/files/overflow-f64-f32.gno:7:13: cannot convert constant of type Float64Kind to Float32Kind