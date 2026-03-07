package main

import "gno.land/r/sys/nameregistry"

var names = []string{}

func main() {
	nameregistry.SetNames(cross, nameregistry.TypeReserved, names...)
}
