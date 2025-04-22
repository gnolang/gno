package main

type Int Int

func main() {}

// Error:
// main/files/recursive8.gno:3:6: invalid recursive type: Int -> Int

// TypeCheckError:
// main/files/recursive8.gno:3:6: invalid recursive type: Int refers to itself
