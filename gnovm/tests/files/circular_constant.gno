package main

const A = B
const B = A + 1

func main() {
}

// Error:
// main/files/circular_constant.gno:3:7: constant definition loop with A

// TypeCheckError:
// main/files/circular_constant.gno:3:7: initialization cycle for A; main/files/circular_constant.gno:3:7: 	A refers to B; main/files/circular_constant.gno:4:7: 	B refers to A
