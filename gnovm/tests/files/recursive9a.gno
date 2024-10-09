package main

type Int = *Int

func main() {}

// Error:
// main/files/recursive9a.gno:1:1: invalid recursive type: Int -> Int