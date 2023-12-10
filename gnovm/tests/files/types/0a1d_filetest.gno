package main

func gen() interface{} {
	return uint64(0)
}
func main() {
	println(0 == gen())

}

// Output:
// false
