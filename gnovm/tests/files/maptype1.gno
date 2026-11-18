package main

// The cycle itself is legal: a map is an indirection and B is a defined
// type, so B resolves to a finite type. B's underlying type is a map
// though, and maps cannot be compared, so B is not a valid map key. It is
// reported at the declaration, once every type in the group is built.

type A = map[B]string

type B A

func main() {
	m := A{}
	println(len(m))
}

// Error:
// main/maptype1.gno:8:6-22: invalid map key type main.B

// TypeCheckError:
// main/maptype1.gno:8:14: invalid map key type B
