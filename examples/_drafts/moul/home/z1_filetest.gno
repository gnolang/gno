package main

import "gno.land/r/moul/home"

func main() {
	println(home.Render(""))
}

// Output:
// Click [here](https://moul.github.io/gno-moul-home-web25/) to visit the full rendering experience.
// # Manfred's (gn)home Dashboard
// ## Meme
// ![meme](https://i.imgflip.com/7ze8dc.jpg)
//
// ## Status
// Online
//
// [update](/r/moul/home$help&func=UpdateStatus)
//
// ## Personal TODO List (bullet list)
// - fill this todo list... [x](/r/moul/home$help&func=DeleteTodo&idx=0)
// - [\[new\]](/r/moul/home$help&func=AddTodo)
// ## Personal TODO List (table)
// | ID | Item | Links |
// | --- | --- | --- |
// | #0 | fill this todo list... | [\[del\]](/r/moul/home$help&func=DeleteTodo&idx=0) |
// ## SVG Example
// this feature may not work with the current gnoweb version and/or configuration.
//
// <svg xmlns="http://www.w3.org/2000/svg" width="500" height="500"><rect x="50" y="50" width="100" height="100" fill="red" /><circle cx="50" cy="50" r="100" fill="red" /><text x="100" y="100" fill="magenta">hello world!</text></svg>
//
// ## Debug
// this feature may not work with the current gnoweb version and/or configuration.
//
// [toggle debug](/r/moul/home:?debug=1)
//
//
