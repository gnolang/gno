package foo

import "github.com/gnolang/gno/_test/foo/boo"

var (
	Bar = "BARR"
	Boo = boo.Boo
)

func init() { println("init foo") }
