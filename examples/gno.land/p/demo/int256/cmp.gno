package int256

func (z *Int) Eq(x *Int) bool {
	return z.value.Eq(&x.value)
}

func (z *Int) Neq(x *Int) bool {
	return !z.Eq(x)
}

// Cmp compares z and x and returns:
//
//   - 1 if z > x
//   - 0 if z == x
//   - -1 if z < x
func (z *Int) Cmp(x *Int) int {
	zSign, xSign := z.Sign(), x.Sign()

	if zSign == xSign {
		return z.value.Cmp(&x.value)
	}

	if zSign == 0 {
		return -xSign
	}

	return zSign
}

// IsZero returns true if z == 0
func (z *Int) IsZero() bool {
	return z.value.IsZero()
}

// IsNeg returns true if z < 0
func (z *Int) IsNeg() bool {
	return z.Sign() < 0
}

func (z *Int) Lt(x *Int) bool {
	return z.Cmp(x) < 0
}

func (z *Int) Gt(x *Int) bool {
	return z.Cmp(x) > 0
}

func (z *Int) Le(x *Int) bool {
	return z.Cmp(x) <= 0
}

func (z *Int) Ge(x *Int) bool {
	return z.Cmp(x) >= 0
}

// Clone creates a new Int identical to z
func (z *Int) Clone() *Int {
	return New().FromUint256(&z.value)
}
