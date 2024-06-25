package main

type (
	nat   []int
	nmap  map[int]int
	nfunc func()
)

func main() {
	var u1 []int
	var n2 nat
	var m map[int]int
	var m2 nmap
	var f func()
	var f2 nfunc

	println(u1)
	println(n2)
	println(m)
	println(m2)
	println(f)
	println(f2)
}

// Output:
// (nil []int)
// (nil main.nat)
// (nil map[int]int)
// (nil main.nmap)
// nil func()()
// (nil main.nfunc)
