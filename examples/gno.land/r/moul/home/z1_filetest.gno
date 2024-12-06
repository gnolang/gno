package main

import "gno.land/r/moul/home"

func main() {
	println(home.Render(""))
}

// Output:
// Click [here](https://ipfs.io/ipfs/12345678) to visit the full rendering experience.\n# Manfred's (gn)home Dashboard
// ## Meme
// ![meme](https://i.imgflip.com/7ze8dc.jpg)
// ## Status
// Online
// [update](/r/moul/home$help&func=UpdateStatus)## Personal ToDo List
// - test
// - fill this todo list... [x](/r/moul/home$help&func=DeleteTodo&idx=0)
// [new](/r/moul/home$help&func=AddTodo)
// ## Table Example
// | foo | bar |
// | --- | --- |
// | foo1 | bar1 |
// | foo2 | bar2 |
//
