package foo

import "filetests/extern/foo/boo"

var (
	Bar = "BARR"
	Boo = boo.Boo
)

func init() { println("init foo") }
