// bitwise contains bitwise operations for Uint instances.
// This file includes functions to perform bitwise AND, OR, XOR, and NOT operations, as well as bit shifting.
// These operations are crucial for manipulating individual bits within a 256-bit unsigned integer.
package uint256

// Or sets z = x | y and returns z.
func (z *Uint) Or(x, y *Uint) *Uint {
	z.arr[0] = x.arr[0] | y.arr[0]
	z.arr[1] = x.arr[1] | y.arr[1]
	z.arr[2] = x.arr[2] | y.arr[2]
	z.arr[3] = x.arr[3] | y.arr[3]
	return z
}

// And sets z = x & y and returns z.
func (z *Uint) And(x, y *Uint) *Uint {
	z.arr[0] = x.arr[0] & y.arr[0]
	z.arr[1] = x.arr[1] & y.arr[1]
	z.arr[2] = x.arr[2] & y.arr[2]
	z.arr[3] = x.arr[3] & y.arr[3]
	return z
}

// Not sets z = ^x and returns z.
func (z *Uint) Not(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = ^x.arr[3], ^x.arr[2], ^x.arr[1], ^x.arr[0]
	return z
}

// AndNot sets z = x &^ y and returns z.
func (z *Uint) AndNot(x, y *Uint) *Uint {
	z.arr[0] = x.arr[0] &^ y.arr[0]
	z.arr[1] = x.arr[1] &^ y.arr[1]
	z.arr[2] = x.arr[2] &^ y.arr[2]
	z.arr[3] = x.arr[3] &^ y.arr[3]
	return z
}

// Xor sets z = x ^ y and returns z.
func (z *Uint) Xor(x, y *Uint) *Uint {
	z.arr[0] = x.arr[0] ^ y.arr[0]
	z.arr[1] = x.arr[1] ^ y.arr[1]
	z.arr[2] = x.arr[2] ^ y.arr[2]
	z.arr[3] = x.arr[3] ^ y.arr[3]
	return z
}

// Lsh sets z = x << n and returns z.
func (z *Uint) Lsh(x *Uint, n uint) *Uint {
	// n % 64 == 0
	if n&0x3f == 0 {
		switch n {
		case 0:
			return z.Set(x)
		case 64:
			return z.lsh64(x)
		case 128:
			return z.lsh128(x)
		case 192:
			return z.lsh192(x)
		default:
			return z.Clear()
		}
	}
	var a, b uint64
	// Big swaps first
	switch {
	case n > 192:
		if n > 256 {
			return z.Clear()
		}
		z.lsh192(x)
		n -= 192
		goto sh192
	case n > 128:
		z.lsh128(x)
		n -= 128
		goto sh128
	case n > 64:
		z.lsh64(x)
		n -= 64
		goto sh64
	default:
		z.Set(x)
	}

	// remaining shifts
	a = z.arr[0] >> (64 - n)
	z.arr[0] = z.arr[0] << n

sh64:
	b = z.arr[1] >> (64 - n)
	z.arr[1] = (z.arr[1] << n) | a

sh128:
	a = z.arr[2] >> (64 - n)
	z.arr[2] = (z.arr[2] << n) | b

sh192:
	z.arr[3] = (z.arr[3] << n) | a

	return z
}

// Rsh sets z = x >> n and returns z.
func (z *Uint) Rsh(x *Uint, n uint) *Uint {
	// n % 64 == 0
	if n&0x3f == 0 {
		switch n {
		case 0:
			return z.Set(x)
		case 64:
			return z.rsh64(x)
		case 128:
			return z.rsh128(x)
		case 192:
			return z.rsh192(x)
		default:
			return z.Clear()
		}
	}
	var a, b uint64
	// Big swaps first
	switch {
	case n > 192:
		if n > 256 {
			return z.Clear()
		}
		z.rsh192(x)
		n -= 192
		goto sh192
	case n > 128:
		z.rsh128(x)
		n -= 128
		goto sh128
	case n > 64:
		z.rsh64(x)
		n -= 64
		goto sh64
	default:
		z.Set(x)
	}

	// remaining shifts
	a = z.arr[3] << (64 - n)
	z.arr[3] = z.arr[3] >> n

sh64:
	b = z.arr[2] << (64 - n)
	z.arr[2] = (z.arr[2] >> n) | a

sh128:
	a = z.arr[1] << (64 - n)
	z.arr[1] = (z.arr[1] >> n) | b

sh192:
	z.arr[0] = (z.arr[0] >> n) | a

	return z
}

// SRsh (Signed/Arithmetic right shift)
// considers z to be a signed integer, during right-shift
// and sets z = x >> n and returns z.
func (z *Uint) SRsh(x *Uint, n uint) *Uint {
	// If the MSB is 0, SRsh is same as Rsh.
	if !x.isBitSet(255) {
		return z.Rsh(x, n)
	}
	if n%64 == 0 {
		switch n {
		case 0:
			return z.Set(x)
		case 64:
			return z.srsh64(x)
		case 128:
			return z.srsh128(x)
		case 192:
			return z.srsh192(x)
		default:
			return z.SetAllOne()
		}
	}
	var a uint64 = MaxUint64 << (64 - n%64)
	// Big swaps first
	switch {
	case n > 192:
		if n > 256 {
			return z.SetAllOne()
		}
		z.srsh192(x)
		n -= 192
		goto sh192
	case n > 128:
		z.srsh128(x)
		n -= 128
		goto sh128
	case n > 64:
		z.srsh64(x)
		n -= 64
		goto sh64
	default:
		z.Set(x)
	}

	// remaining shifts
	z.arr[3], a = (z.arr[3]>>n)|a, z.arr[3]<<(64-n)

sh64:
	z.arr[2], a = (z.arr[2]>>n)|a, z.arr[2]<<(64-n)

sh128:
	z.arr[1], a = (z.arr[1]>>n)|a, z.arr[1]<<(64-n)

sh192:
	z.arr[0] = (z.arr[0] >> n) | a

	return z
}

func (z *Uint) lsh64(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = x.arr[2], x.arr[1], x.arr[0], 0
	return z
}

func (z *Uint) lsh128(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = x.arr[1], x.arr[0], 0, 0
	return z
}

func (z *Uint) lsh192(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = x.arr[0], 0, 0, 0
	return z
}

func (z *Uint) rsh64(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = 0, x.arr[3], x.arr[2], x.arr[1]
	return z
}

func (z *Uint) rsh128(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = 0, 0, x.arr[3], x.arr[2]
	return z
}

func (z *Uint) rsh192(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = 0, 0, 0, x.arr[3]
	return z
}

func (z *Uint) srsh64(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = MaxUint64, x.arr[3], x.arr[2], x.arr[1]
	return z
}

func (z *Uint) srsh128(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = MaxUint64, MaxUint64, x.arr[3], x.arr[2]
	return z
}

func (z *Uint) srsh192(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = MaxUint64, MaxUint64, MaxUint64, x.arr[3]
	return z
}
