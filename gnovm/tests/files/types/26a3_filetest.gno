package main

func gen() interface{} {
	return false
}

func main() {
	gen()++
}
