package main

import (
	"gno.land/r/x/jeronimo_render_proxy/home"
	_ "gno.land/r/x/jeronimo_render_proxy/home/v1"
	_ "gno.land/r/x/jeronimo_render_proxy/home/v2"
)

func main() {
	println(home.CurrentRealmPath())
	println(home.Render(""))
}

// Output:
// gno.land/r/x/jeronimo_render_proxy/home/v2
// Rendered by v2
