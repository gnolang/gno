package main

// A use or assignment before a shadowing const declaration in the same block
// refers to the outer name, which is not const.
func main() {
	v := 1
	{
		println(v)
		const v = "c"
		println(v)
	}
	w := 1
	{
		w++
		w += 2
		println(w)
		const w = "k"
		println(w)
	}
	println(v, w)
}

// Output:
// 1
// c
// 4
// k
// 1 4
