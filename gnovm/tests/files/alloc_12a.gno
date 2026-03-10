// PKGPATH: gno.land/r/test
// MAXALLOC: 10000
package test

import (
	"runtime"
)

var s string

func Assign() {
	s = `aaaaaaaaaaaaaaaaaaaa`
}

func init(){
	Assign()
}

func main() {
	s1 := s
	// println("===before GC, MemStats: ", runtime.MemStats())
	runtime.GC()
	// println("===after GC, MemStats: ", runtime.MemStats())
}

// TypeCheckError:
// gno.land/r/test/alloc_12a.gno:20:2: declared and not used: s1
