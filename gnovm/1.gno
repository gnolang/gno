package main

var arr1 = [2]int{1, 2}
var arr2 = [2]int{1, 2}

//func init(){

//}

func main() {
	println(&arr1[0] == &arr2[0])
}
