package steal_ownership

var (
	ptr   *uint32
	slice []uint32
)

func Steal(xptr *uint32) {
	ptr = xptr
}

func StealSlice(xslice []uint32) {
	slice = xslice
}
