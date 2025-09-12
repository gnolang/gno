package int256

// Not sets z to the bitwise NOT of x and returns z.
//
// The bitwise NOT operation flips each bit of the operand.
func (z *Int) Not(x *Int) *Int {
	z.value.Not(&x.value)
	return z
}

// And sets z to the bitwise AND of x and y and returns z.
//
// The bitwise AND operation results in a value that has a bit set
// only if both corresponding bits of the operands are set.
func (z *Int) And(x, y *Int) *Int {
	z.value.And(&x.value, &y.value)
	return z
}

// Or sets z to the bitwise OR of x and y and returns z.
//
// The bitwise OR operation results in a value that has a bit set
// if at least one of the corresponding bits of the operands is set.
func (z *Int) Or(x, y *Int) *Int {
	z.value.Or(&x.value, &y.value)
	return z
}

// Xor sets z to the bitwise XOR of x and y and returns z.
//
// The bitwise XOR operation results in a value that has a bit set
// only if the corresponding bits of the operands are different.
func (z *Int) Xor(x, y *Int) *Int {
	z.value.Xor(&x.value, &y.value)
	return z
}

// Rsh sets z to the result of right-shifting x by n bits and returns z.
//
// Right shift operation moves all bits in the operand to the right by the specified number of positions.
// Bits shifted out on the right are discarded, and zeros are shifted in on the left.
func (z *Int) Rsh(x *Int, n uint) *Int {
	z.value.Rsh(&x.value, n)
	return z
}

// Lsh sets z to the result of left-shifting x by n bits and returns z.
//
// Left shift operation moves all bits in the operand to the left by the specified number of positions.
// Bits shifted out on the left are discarded, and zeros are shifted in on the right.
func (z *Int) Lsh(x *Int, n uint) *Int {
	z.value.Lsh(&x.value, n)
	return z
}
