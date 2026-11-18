// MAXALLOC: 50000
// PKGPATH: gno.land/r/test

// VM panic text is untracked until recover hands it to user code. Here
// it embeds a 1000-byte key; once held, the GC recount must include it.
package test

import (
	"chain/params"
	"runtime"
)

var held any

func grab(key string) {
	defer func() { held = recover() }()
	params.GetString(key) // panics "invalid param key: " + key
}

func main() {
	b := make([]byte, 1000)
	for i := range b {
		b[i] = ':'
	}
	grab(string(b))
	runtime.GC()
	println(len(held.(string)))
	println(runtime.MemStats())
}

// Output:
// 1019
// Allocator{maxBytes:50000, bytes:14611}
