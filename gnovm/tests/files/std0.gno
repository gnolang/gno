package main

func main() {
	println(20)
	name := "test1"
	if len(name) > 20 {
		panic("should not happen")
	}
	println("done")
}

// Output:
// 20
// done
