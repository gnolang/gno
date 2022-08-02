package atomic

// XXX shims.

func CompareAndSwapInt32(ptr *int32, old, new_ int32) {
	if *ptr == old {
		*ptr = new_
	}
}

func AddInt32(ptr *int32, x int32) {
	*ptr += x
}
