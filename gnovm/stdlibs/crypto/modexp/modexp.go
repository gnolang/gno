package modexp

import "math/big"

// X_modExp mirrors EIP-198's MODEXP precompile.
func X_modExp(base, exp, modulus []byte) []byte {
	out := make([]byte, len(modulus))
	m := new(big.Int).SetBytes(modulus)
	if m.Sign() == 0 {
		return out
	}
	b := new(big.Int).SetBytes(base)
	e := new(big.Int).SetBytes(exp)
	r := new(big.Int).Exp(b, e, m)
	r.FillBytes(out)
	return out
}
