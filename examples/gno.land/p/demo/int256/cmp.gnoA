package int256

// Eq returns true if z == x
func (z *Int) Eq(x *Int) bool {
	return (z.neg == x.neg) && z.abs.Eq(x.abs)
}

// Neq returns true if z != x
func (z *Int) Neq(x *Int) bool {
	return !z.Eq(x)
}

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
func (z *Int) Cmp(x *Int) (r int) {
	// x cmp y == x cmp y
	// x cmp (-y) == x
	// (-x) cmp y == y
	// (-x) cmp (-y) == -(x cmp y)
	switch {
	case z == x:
		// nothing to do
	case z.neg == x.neg:
		r = z.abs.Cmp(x.abs)
		if z.neg {
			r = -r
		}
	case z.neg:
		r = -1
	default:
		r = 1
	}
	return
}

// IsZero returns true if z == 0
func (z *Int) IsZero() bool {
	return z.abs.IsZero()
}

// IsNeg returns true if z < 0
func (z *Int) IsNeg() bool {
	return z.neg
}

// Lt returns true if z < x
func (z *Int) Lt(x *Int) bool {
	if z.neg {
		if x.neg {
			return z.abs.Gt(x.abs)
		} else {
			return true
		}
	} else {
		if x.neg {
			return false
		} else {
			return z.abs.Lt(x.abs)
		}
	}
}

// Gt returns true if z > x
func (z *Int) Gt(x *Int) bool {
	if z.neg {
		if x.neg {
			return z.abs.Lt(x.abs)
		} else {
			return false
		}
	} else {
		if x.neg {
			return true
		} else {
			return z.abs.Gt(x.abs)
		}
	}
}

// Clone creates a new Int identical to z
func (z *Int) Clone() *Int {
	return &Int{z.abs.Clone(), z.neg}
}
