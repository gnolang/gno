package main

type Int = []Int

func main() {}

// Error:
// main/recursive9c.gno:3:6-17: invalid recursive type: Int -> Int

// TypeCheckError:
// main/recursive9c.gno:3:6: invalid recursive type: Int refers to itself
