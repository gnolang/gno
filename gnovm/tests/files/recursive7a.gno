package main

type S [2]S

func main() {}

// Error:
// main/recursive7a.gno:3:6-12: invalid recursive type: S -> S

// TypeCheckError:
// main/recursive7a.gno:3:6: invalid recursive type: S refers to itself
