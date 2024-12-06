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
	println(home.Render("?debug=1"))
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
// ## SVG Example
// <svg xmlns="http://www.w3.org/2000/svg" width="500" height="500"><rect x="50" y="50" width="100" height="100" fill="red" /><circle cx="50" cy="50" r="100" fill="red" /><text x="100" y="100" fill="magenta">hello world!</text></svg>
// <details><summary>debug</summary>
//
// ### Logs
// - hello world!
// ### Metadata
// | Key | Value |
// | --- | --- |
// | `std.CurrentRealm().PkgPath()` | gno.land/r/moul/home |
// | `std.CurrentRealm().Addr()` | g1h8h57ntxadcze3f703skymfzdwa6t3ugf0nq3z |
// | `std.GetHeight()` | 123 |
//
// </details>
//
