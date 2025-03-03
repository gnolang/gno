package gc

import "runtime"

// TODO: Refactor this logic to handle a wider range of cases
func Alloc1() {
	data := make([]byte, 100*1024*1024)
	data = nil
	runtime.GC()
	println("MemStats after GC: ", runtime.MemStats())
}
