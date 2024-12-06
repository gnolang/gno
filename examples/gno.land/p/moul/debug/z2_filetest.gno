package main

import "gno.land/p/moul/debug"

func main() {
	var d debug.Debug
	d.Log("hello world!")
	d.Log("foobar")
	println("---")
	println(d.Render(""))
	println("---")
	println(d.Render("?debug=1"))
	println("---")
}

// Output:
// ---
//
// ---
// <details><summary>debug</summary>
//
// ### Logs
// - hello world!
// - foobar
// ### Metadata
// | Key | Value |
// | --- | --- |
// | `std.CurrentRealm().PkgPath()` |  |
// | `std.CurrentRealm().Addr()` | g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm |
// | `std.GetHeight()` | 123 |
//
// </details>
//
// ---
