/*
Package ascii provides a simple way to create ASCII components in Gno.

The ascii package allows you to create basic ASCII components like boxes and lines, and output the generated ASCII to a string in codeblock format when necessary.

Example:

	import "gno.land/p/lou/ascii"

	func Foo() string {
	    return ascii.Frame([]string{"Hello, World!"}, 20, 5, ascii.AlignWCenter, ascii.AlignHCenter)
	}

*/

package ascii // import "gno.land/p/lou/ascii"
