// MAXALLOC: 5000000
package main

// A declared key type over a primitive — the map[address]T shape that
// dominates gno.land realms. The key's own bytes are short, so the omitted
// TypeID prefix ("main.Address") is a large fraction of the charge.

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
// 125449
