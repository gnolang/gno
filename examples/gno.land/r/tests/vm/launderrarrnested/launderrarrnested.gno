package launderrarrnested

import "strconv"

type N [2][3]int
type S struct{ Arr [3]int }
type P struct{ Buf [4]byte }

var (
	gN *N
	gS *S
	gP *P
)

// Each Register validates the value once and then trusts it.
func RegisterN(n *N) bool {
	if n[0][0] != 0 {
		return false
	}
	gN = n
	return true
}

func RegisterS(s *S) bool {
	if s.Arr[0] != 0 {
		return false
	}
	gS = s
	return true
}

func RegisterP(p *P) bool {
	if p.Buf[0] != 0 {
		return false
	}
	gP = p
	return true
}

func Trusted() string {
	return strconv.Itoa(gN[0][0]) + " " + strconv.Itoa(gS.Arr[0]) + " " + strconv.Itoa(int(gP.Buf[0]))
}

func SetS(cur realm, v int) { gS.Arr[0] = v }
