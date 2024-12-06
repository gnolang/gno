package main

import (
	"std"

	"gno.land/r/moul/home"
)

func main() {
	std.TestSetOrigCaller("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
	home.AddTodo("aaa")
	home.AddTodo("bbb")
	home.AddTodo("ccc")
	home.AddTodo("ddd")
	home.AddTodo("eee")
	home.UpdateStatus("Lorem Ipsum")
	home.DeleteTodo(3)
	println(home.Render(""))
}

// Output:
// Click [here](https://ipfs.io/ipfs/12345678) to visit the full rendering experience.\n# Manfred's (gn)home Dashboard
// ## Meme
// ![meme](https://i.imgflip.com/7ze8dc.jpg)
// ## Status
// Lorem Ipsum
// [update](/r/moul/home$help&func=UpdateStatus)## Personal ToDo List
// - test
// - fill this todo list... [x](/r/moul/home$help&func=DeleteTodo&idx=0)
// - aaa [x](/r/moul/home$help&func=DeleteTodo&idx=1)
// - bbb [x](/r/moul/home$help&func=DeleteTodo&idx=2)
// - ddd [x](/r/moul/home$help&func=DeleteTodo&idx=3)
// - eee [x](/r/moul/home$help&func=DeleteTodo&idx=4)
// [new](/r/moul/home$help&func=AddTodo)
// ## Table Example
// | foo | bar |
// | --- | --- |
// | foo1 | bar1 |
// | foo2 | bar2 |
//
