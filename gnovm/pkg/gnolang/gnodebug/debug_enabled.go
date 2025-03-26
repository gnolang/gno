//go:build debug

package gnodebug

import (
	"os"
	"sync"
)

const Debug DebugType = true

var flagsOnce = sync.OnceValue[DebugFlags](func() DebugFlags {
	return ParseFlags(os.Getenv("GNODEBUG"))
})

func (DebugType) Printf(kind, format string, args ...any) {
	flagsOnce().Printf(kind, format, args...)
}

func (DebugType) Get(flagName string) string {
	return flagsOnce()[flagName]
}

func (DebugType) Enabled(flagName string) bool {
	return flagsOnce()[flagName] == "1"
}

func (DebugType) Set(flagName, val string) {
	flagsOnce()[flagName] = val
}
