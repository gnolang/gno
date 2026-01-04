package main

type Int = func() Int

func main() {}

// Error:
// main/recursive9b.gno:3:6-22: invalid recursive type: Int -> Int

// TypeCheckError:
// main/recursive9b.gno:3:6: invalid recursive type: Int refers to itself
