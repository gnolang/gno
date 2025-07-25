package main

type Int = Int

func main() {}

// Error:
// main/recursive9.gno:3:6-15: invalid recursive type: Int -> Int

// TypeCheckError:
// main/recursive9.gno:3:6: invalid recursive type: Int refers to itself
