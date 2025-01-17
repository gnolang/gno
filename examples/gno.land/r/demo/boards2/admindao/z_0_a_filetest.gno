// PKGPATH: gno.land/r/demo/boards2/test
package test

import (
	"std"

	"gno.land/r/demo/boards2/admindao"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "test1"
)

func init() {
	std.TestSetOrigCaller(owner)
}

func main() {
	dao := admindao.New(name)
	println("ok")

	if admindao.Exists(name) {
		println("exists")
	}

	if got, found := admindao.Get(name); found {
		if got == dao {
			print("found")
		}
	}
}

// Output:
// ok
// exists
// found
