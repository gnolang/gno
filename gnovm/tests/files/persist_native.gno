package main

var a any

func init() {
	a = println
}

func main() {
	a("works")
}
