package int256

func (z *Int) Not(x *Int) *Int {
	z.value.Not(&x.value)
	return z
}

func (z *Int) And(x, y *Int) *Int {
	z.value.And(&x.value, &y.value)
	return z
}

func (z *Int) Or(x, y *Int) *Int {
	z.value.Or(&x.value, &y.value)
	return z
}

func (z *Int) Xor(x, y *Int) *Int {
	z.value.Xor(&x.value, &y.value)
	return z
}

func (z *Int) Rsh(x *Int, n uint) *Int {
	z.value.Rsh(&x.value, n)
	return z
}

func (z *Int) Lsh(x *Int, n uint) *Int {
	z.value.Lsh(&x.value, n)
	return z
}
