package main

var i = new(int)

func init() {
	println(i == i)
}

func main() {
	*i = 1
	println(i == i)
}
