package main

type Int Int

func main() {}

// Error:
// main/recursive8.gno:3:6-13: invalid recursive type: Int -> Int

// TypeCheckError:
// main/recursive8.gno:3:6: invalid recursive type: Int refers to itself
