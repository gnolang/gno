package main

import (
	"std"

	"gno.land/r/moul/home"
)

func main() {
	std.TestSetOrigCaller("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
	home.AddNewTodo("aaa")
	home.AddNewTodo("bbb")
	home.AddNewTodo("ccc")
	home.AddNewTodo("ddd")
	home.AddNewTodo("eee")
	home.UpdateStatus("Lorem Ipsum")
	home.DeleteTodo(3)
	println(home.Render(""))
}

// Output:
// Checkout the full experience from the [web2.5 frontend](https://ipfs.io/ipfs/12345678)# Manfred's (gn)home Dashboard
// ## Meme
// ![meme](https://i.imgflip.com/7ze8dc.jpg)
// ## Status
// Lorem Ipsum
// [update](/r/moul/home$help&func=UpdateStatus)## Personal ToDo List
// - test
// - fill this todo list... [del](/r/moul/home$help&func=DeleteTodo&idx=0)
// - aaa [del](/r/moul/home$help&func=DeleteTodo&idx=1)
// - bbb [del](/r/moul/home$help&func=DeleteTodo&idx=2)
// - ddd [del](/r/moul/home$help&func=DeleteTodo&idx=3)
// - eee [del](/r/moul/home$help&func=DeleteTodo&idx=4)
// [new todo](/r/moul/home$help&func=AddNewTodo)
//
