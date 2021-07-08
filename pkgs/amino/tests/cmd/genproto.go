package main

import (
	"github.com/gnolang/gno/pkgs/amino/genproto"
	"github.com/gnolang/gno/pkgs/amino/tests"
)

func main() {
	pkg := tests.Package
	genproto.WriteProto3Schema(pkg)
	genproto.MakeProtoFolder(pkg, "proto")
	genproto.RunProtoc(pkg, "proto")
	genproto.WriteProtoBindings(pkg)
}
