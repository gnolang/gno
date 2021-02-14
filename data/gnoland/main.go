package main

var root Node

type Node interface{}
type Key interface{}
type Value interface{}

type InnerNode struct {
	Key   Key
	Left  Node `gno:owned`
	Right Node `gno:owned`
}

type LeafNode struct {
	Key   Key
	Value Value
}

func init() {
	root = InnerNode{
		Key: "old",
	}
}

func main() {
	println("gno.land initializing...")
	node1 := LeafNode{
		Key:   "left",
		Value: "left value",
	}
	node2 := LeafNode{
		Key:   "right",
		Value: "right value",
	}
	root = InnerNode{
		Key:   "new",
		Left:  node1,
		Right: node2,
	}
	println("gno.land ready!")
}
