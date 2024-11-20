package stdlibs

import "embed"

// embedded embeds the stdlibs.
// Be careful to remove transpile artifacts before building release binaries or they will be included
//
//go:embed */*
var embedded embed.FS

func Embedded() embed.FS {
	return embedded
}
