package main

func main() {
    const b4 = uint64(1) << 65
}

// Error:
// main/overflow14.gno:3:1-5:2: constant overflows

// TypeCheckError:
// main/overflow14.gno:4:16: uint64(1) << 65 (constant 36893488147419103232 of type uint64) overflows uint64
