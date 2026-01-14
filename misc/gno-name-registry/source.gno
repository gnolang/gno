package main

import (
	"strings"

	"gno.land/r/sys/nameregistry"
)

var entries = []struct {
	Name string
	Type int // 0: full, 1: TLD, 2: custom
}{}

func main() {
	names := make([]string, len(entries))
	for _, e := range entries {
		i := strings.Index(e.Name, ".")
		if i == -1 {
			names = append(names, e.Name)
		} else {
			names = append(names, e.Name[:i])
		}
	}

	nameregistry.SetNames(cross, nameregistry.TypeReserved, names...)
}
