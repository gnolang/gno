package main

import "gno.land/r/manfred/home"

func main() {
	std.TestSetOrigCaller(config.Addr())
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
// Online
//
// ## Personal ToDo List
// - [ ] fill this todo list...
