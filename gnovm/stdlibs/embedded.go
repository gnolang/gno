package stdlibs

import "embed"

// embeddedSources embeds the stdlibs.
// Be careful to remove transpile artifacts before building release binaries or they will be included.
//
//go:embed */*
var embeddedSources embed.FS

// EmbeddedSources returns embedded stdlibs sources.
func EmbeddedSources() embed.FS {
	return embeddedSources
}
