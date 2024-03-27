package int256

func (z *Int) Add(x, y *Int) *Int {
	z.initiateAbs()

	neg := x.neg

	if x.neg == y.neg {
		// x + y == x + y
		// (-x) + (-y) == -(x + y)
		z.abs = z.abs.Add(x.abs, y.abs)
	} else {
		// x + (-y) == x - y == -(y - x)
		// (-x) + y == y - x == -(x - y)
		if x.abs.Cmp(y.abs) >= 0 {
			z.abs = z.abs.Sub(x.abs, y.abs)
		} else {
			neg = !neg
			z.abs = z.abs.Sub(y.abs, x.abs)
		}
	}
	z.neg = neg // 0 has no sign
	return z
}

// Sub sets z to the difference x-y and returns z.
func (z *Int) Sub(x, y *Int) *Int {
	z.initiateAbs()

	neg := x.neg
	if x.neg != y.neg {
		// x - (-y) == x + y
		// (-x) - y == -(x + y)
		z.abs = z.abs.Add(x.abs, y.abs)
	} else {
		// x - y == x - y == -(y - x)
		// (-x) - (-y) == y - x == -(x - y)
		if x.abs.Cmp(y.abs) >= 0 {
			z.abs = z.abs.Sub(x.abs, y.abs)
		} else {
			neg = !neg
			z.abs = z.abs.Sub(y.abs, x.abs)
		}
	}
	z.neg = neg // 0 has no sign
	return z
}

// Mul sets z to the product x*y and returns z.
func (z *Int) Mul(x, y *Int) *Int {
	z.initiateAbs()

	z.abs = z.abs.Mul(x.abs, y.abs)
	z.neg = x.neg != y.neg // 0 has no sign
	return z
}

func (z *Int) Div(x, y *Int) *Int {
	z.initiateAbs()

	z.abs.Div(x.abs, y.abs)
	if x.neg == y.neg {
		z.neg = false
	} else {
		z.neg = true
	}
	return z
}

// Quo sets z to the quotient x/y for y != 0 and returns z.
// If y == 0, a division-by-zero run-time panic occurs.
// Quo implements truncated division (like Go); see QuoRem for more details.
func (z *Int) Quo(x, y *Int) *Int {
	z.initiateAbs()

	z.abs = z.abs.Div(x.abs, y.abs)
	z.neg = len(z.abs) > 0 && x.neg != y.neg // 0 has no sign
	return z
}

// Rem sets z to the remainder x%y for y != 0 and returns z.
// If y == 0, a division-by-zero run-time panic occurs.
// Rem implements truncated modulus (like Go); see QuoRem for more details.
func (z *Int) Rem(x, y *Int) *Int {
	z.initiateAbs()

	z.abs.Mod(x.abs, y.abs)
	z.neg = z.abs.Sign() > 0 && x.neg // 0 has no sign
	return z
}
