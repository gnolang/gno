package main

var arr1 [2]int
var arr2 [2]int

func init() {
	arr1 = [2]int{1, 2}
	arr2 = [2]int{1, 2}
}

func main() {
	println(&arr1 == &arr1)
}
