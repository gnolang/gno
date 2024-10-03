package main

type Int = func() Int

func main() {}

// Error:
// main/files/recursive9b.gno:1:1: invalid recursive type: Int -> Int