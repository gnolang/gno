package main

import (
	"github.com/gnolang/gno/gnovm/pkg/sogno"
	"github.com/tidwall/btree"
)

var MyVar int64 = 2

func main() {
	bmap := btree.NewMap[string, interface{}](2)
	bmap.Set("MyVar", &MyVar)
	ctx := &sogno.Context{
		Path:  "gno.land/r/demo/users",
		State: bmap,
	}
	ctx.Main()
}
