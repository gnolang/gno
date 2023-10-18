package main

import "gno.land/r/manfred/present"

func main() {
	println(present.Render(""))
	println("------------------------------------")
	println(present.Render("p/miami23"))
}

// Output:
// <div class='columns-3'><div>
//
// ## [Portal Loop Demo (Miami 2023)](/r/manfred/present:p/miami23)
// **[Learn More](/r/manfred/present:p/miami23)**
//
// </div></div>
// ------------------------------------
// # Portal Loop Demo (Miami 2023)
//
// - roadmap
// - portal loop
// - screenshot github project
//
// [#demo](/r/manfred/present:t/demo) [#portal-loop](/r/manfred/present:t/portal-loop) [#miami](/r/manfred/present:t/miami)
//
// by g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq on 1970-01-01 12:00am UTC
