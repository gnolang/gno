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
// Click [here](https://ipfs.io/ipfs/12345678) to visit the full rendering experience.
// # Manfred's (gn)home Dashboard
// ## Meme
// ![meme](https://i.imgflip.com/7ze8dc.jpg)
// ## Status
// Lorem Ipsum
// [update](/r/moul/home$help&func=UpdateStatus)
// ## Personal TODO List (bullet list)
// - test
// - fill this todo list... [x](/r/moul/home$help&func=DeleteTodo&idx=0)
// - aaa [x](/r/moul/home$help&func=DeleteTodo&idx=1)
// - bbb [x](/r/moul/home$help&func=DeleteTodo&idx=2)
// - ddd [x](/r/moul/home$help&func=DeleteTodo&idx=3)
// - eee [x](/r/moul/home$help&func=DeleteTodo&idx=4)
// [new](/r/moul/home$help&func=AddTodo)
// ## Personal TODO List (table)
// | ID | Item | Links |
// | --- | --- | --- |
// | #0 | fill this todo list... | [del](/r/moul/home$help&func=DeleteTodo&idx=0) |
// | #1 | aaa | [del](/r/moul/home$help&func=DeleteTodo&idx=1) |
// | #2 | bbb | [del](/r/moul/home$help&func=DeleteTodo&idx=2) |
// | #3 | ddd | [del](/r/moul/home$help&func=DeleteTodo&idx=3) |
// | #4 | eee | [del](/r/moul/home$help&func=DeleteTodo&idx=4) |
//
