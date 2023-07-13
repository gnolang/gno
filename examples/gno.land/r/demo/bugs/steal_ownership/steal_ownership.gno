package steal_ownership

var ptr *uint32

func Steal(xptr *uint32) {
	ptr = xptr
}
