package main

func main() {
	var i interface{} = [2]int{0, 1}
	const a = len(i.([2]int))
	println("ok", a)
}

// Output:
// ok 2