package rat

//----------------------------------------
// Rat fractions

// represents a fraction.
type Rat struct {
	X int32
	Y int32 // must be positive
}

func NewRat(x, y int32) Rat {
	if y <= 0 {
		panic("invalid std.Rat denominator")
	}
	return Rat{X: x, Y: y}
}

func (r1 Rat) IsValid() bool {
	if r1.Y <= 0 {
		return false
	}
	return true
}

func (r1 Rat) Cmp(r2 Rat) int {
	if !r1.IsValid() {
		panic("invalid std.Rat left operand")
	}
	if !r2.IsValid() {
		panic("invalid std.Rat right operand")
	}
	var p1, p2 int64
	p1 = int64(r1.X) * int64(r2.Y)
	p2 = int64(r1.Y) * int64(r2.X)
	if p1 < p2 {
		return -1
	} else if p1 == p2 {
		return 0
	} else {
		return 1
	}
}

//func (r1 Rat) Plus(r2 Rat) Rat {
// XXX
//}
