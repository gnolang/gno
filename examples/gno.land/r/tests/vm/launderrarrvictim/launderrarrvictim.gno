package launderrarrvictim

type A [3]int
type B [3]byte

var (
	gA     *A
	gB     *B
	gSlice *A
)

func Register(a *A) bool {
	if a[0] != 0 {
		return false
	}
	gA = a
	return true
}

func RegisterBytes(b *B) bool {
	if b[0] != 0 {
		return false
	}
	gB = b
	return true
}

func RegisterSlice(a *A) bool {
	if a[0] != 0 {
		return false
	}
	gSlice = a
	return true
}

func Trusted() int       { return gA[0] }
func TrustedBytes() byte { return gB[0] }
func TrustedSlice() int  { return gSlice[0] }

func VictimWrite(cur realm, a int, b byte, slice int) {
	gA[0] = a
	gB[0] = b
	gSlice[0] = slice
}
