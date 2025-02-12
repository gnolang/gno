package uint256

func (z *Uint) NilToZero() *Uint {
	if z == nil {
		z = NewUint(0)
	}
	return z
}
