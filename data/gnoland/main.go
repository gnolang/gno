package main

var root Node

type Node interface{}
type Key interface{}

type InnerNode struct {
	Key   Key
	Left  Node `gno:owned`
	Right Node `gno:owned`
}

func init() {
	root = InnerNode{
		Key: "old",
	}
}

func main() {
	println("gno.land initializing...")
	root = InnerNode{
		Key: "new",
	}
	println("gno.land ready!")
}
