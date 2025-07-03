package main

type Int = *Int

func main() {}

// Error:
// main/recursive9a.gno:3:6-16: invalid recursive type: Int -> Int

// TypeCheckError:
// main/recursive9a.gno:3:6: invalid recursive type: Int refers to itself
