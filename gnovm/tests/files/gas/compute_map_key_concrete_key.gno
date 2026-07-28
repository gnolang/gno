// MAXALLOC: 5000000
package main

// A declared key type over a primitive — the map[address]T shape that
// dominates gno.land realms. The omitted TypeID prefix ("main.Address")
// is ~29% of the bytes hashed but only ~6 of the 98 gas charged per access.

type Address string

func main() {
	m := map[Address]int{}
	for i := 0; i < 100; i++ {
		m["g1qwertyuiopasdfghjklzxcvbnm0000"] = i
	}
	println(len(m))
}

// Output:
// 1

// Gas:
// 124849
