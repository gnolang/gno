// MAXALLOC: 50000
package main

import "runtime"

type littleEndian struct{}
type bigEndian struct{}

func (littleEndian) PutUint32(b []byte, v uint32) {
	// do nothing
	println("memstats in littleEndian before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in littleEndian after GC: ", runtime.MemStats())
}

func (*bigEndian) PutUint32(b []byte, v uint32) {
	// do nothing
	println("memstats in *bigEndian before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in *bigEndian after GC: ", runtime.MemStats())
}

var LittleEndian littleEndian
var BigEndian = &bigEndian{}
var BigEndian2 bigEndian

func main() {
	LittleEndian.PutUint32([]byte("a"), 1)
	BigEndian.PutUint32([]byte("b"), 1)
	BigEndian2.PutUint32([]byte("b"), 1)
	println("memstats in main before GC: ", runtime.MemStats())
	runtime.GC()
	println("memstats in main after GC: ", runtime.MemStats())
}

// Output:
// memstats in littleEndian before GC:  Allocator{maxBytes:50000, bytes:4280}
// memstats in littleEndian after GC:  Allocator{maxBytes:50000, bytes:3768}
// memstats in *bigEndian before GC:  Allocator{maxBytes:50000, bytes:6541}
// memstats in *bigEndian after GC:  Allocator{maxBytes:50000, bytes:4560}
// memstats in *bigEndian before GC:  Allocator{maxBytes:50000, bytes:7589}
// memstats in *bigEndian after GC:  Allocator{maxBytes:50000, bytes:4024}
// memstats in main before GC:  Allocator{maxBytes:50000, bytes:5517}
// memstats in main after GC:  Allocator{maxBytes:50000, bytes:2776}
