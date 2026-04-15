// PKGPATH: gno.land/r/test
// MAXALLOC: 200000
package test

import (
	"runtime"
	"strings"
)

var s string

func Assign() {
	s = strings.Repeat("a", 200000)
}

func init(){
	Assign()
}

func main() {
	s1 := s
	println("===before GC, MemStats: ", runtime.MemStats())
	runtime.GC()
	println("===after GC, MemStats: ", runtime.MemStats())
}

// TypeCheckError:
// gno.land/r/test/alloc_13b.gno:21:2: declared and not used: s1

// Error:
// allocation limit exceeded
