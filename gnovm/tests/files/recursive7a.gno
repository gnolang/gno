package main

type S [2]S

func main() {}

// Error:
// main/files/recursive7a.gno:3:6: invalid recursive type: S -> S

// TypeCheckError:
// main/files/recursive7a.gno:3:6: invalid recursive type: S refers to itself
