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
// # Manfred's (gn)home Dashboard
//
// ## Meme
// ![](https://i.imgflip.com/7ze8dc.jpg)
//
// ## Status
// Lorem Ipsum
//
// ## Personal ToDo List
// - [ ] fill this todo list...
// - [ ] aaa
// - [ ] bbb
// - [ ] ddd
// - [ ] eee
