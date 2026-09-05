package main

// The cycle itself is legal: a map is an indirection and B is a defined
// type, so B resolves to a finite type. B's underlying type is a map
// though, and maps cannot be compared, so B is not a valid map key.

type A = map[B]string

type B A

func main() {}

// Error:
// main/maptype1.gno:7:6-22: invalid map key type main.B

// TypeCheckError:
// main/maptype1.gno:7:14: invalid map key type B
