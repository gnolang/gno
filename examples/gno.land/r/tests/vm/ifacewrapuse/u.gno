package ifacewrapuse

import (
	"gno.land/p/demo/tests/ifacewrap"
	"gno.land/p/demo/tests/launderpkg/v0"
)

var gO = &launderpkg.Object{Field: "orig"}

type Putter interface{ Put(string) }

func Direct() string { ifacewrap.New(gO).Put("direct"); return gO.Field }
func Iface() string  { var p Putter = ifacewrap.New(gO); p.Put("iface"); return gO.Field }
func IfaceV() string { var p Putter = ifacewrap.WrapV{gO}; p.Put("ifacev"); return gO.Field }
