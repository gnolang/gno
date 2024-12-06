package main

import "gno.land/r/moul/home"

func main() {
	println(home.Render(""))
}

// Output:
// Click [here](https://ipfs.io/ipfs/12345678) to visit the full rendering experience.
// # Manfred's (gn)home Dashboard
// ## Meme
// ![meme](https://i.imgflip.com/7ze8dc.jpg)
// ## Status
// Online
// [update](/r/moul/home$help&func=UpdateStatus)
// ## Personal TODO List (bullet list)
// - test
// - fill this todo list... [x](/r/moul/home$help&func=DeleteTodo&idx=0)
// [new](/r/moul/home$help&func=AddTodo)
// ## Personal TODO List (table)
// | ID | Item | Links |
// | --- | --- | --- |
// | #0 | fill this todo list... | [del](/r/moul/home$help&func=DeleteTodo&idx=0) |
//
