package main

import "errors"

var e1 = errors.New("EOF")
var e2 = errors.New("EOF")

func main() {

	println(e1 == e2)

}
