package main

type S [2]S

func main() {}

// Error:
// main/files/recursive7a.gno:1:1: invalid recursive type: S -> S
