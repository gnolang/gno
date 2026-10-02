package main

var myVar string = "default"

const constVar = "unpatchable"

func main() {
	println(myVar)
	_ = constVar
}
