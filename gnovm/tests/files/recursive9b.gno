package main

type Int = func() Int

func main() {}

// Error:
// main/files/recursive9b.gno:3:6: invalid recursive type: Int -> Int

// TypeCheckError:
// main/files/recursive9b.gno:3:6: invalid recursive type: Int refers to itself
