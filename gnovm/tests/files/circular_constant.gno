package main

const A = B
const B = A + 1

func main() {
}

// Error:
// main/files/circular_constant.gno:1: constant definition loop with A
