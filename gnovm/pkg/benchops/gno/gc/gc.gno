package gc

import "runtime"

func Alloc1() {
	data := make([]byte, 100*1024*1024)
	data = nil
	runtime.GC()
	println("MemStats after GC: ", runtime.MemStats())
}
