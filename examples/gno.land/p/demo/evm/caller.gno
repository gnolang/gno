package evm

import (
	"std"

	"gno.land/p/demo/uint256"
)

type Caller struct {
	EVMAddress *uint256.Uint
	GnoAddress std.Address
	Gas uint
}

func (c *Caller) Call(env *Env, gasAmount uint, code []byte) error {
	if gasAmount > c.Gas {
		return notEnoughGas
	}
	st := newState()
	st.setEnv(env)
	st.setGasBet(gasAmount)
	err := st.executeByteCode(code)
	if err != nil {
		return err
	}
	return nil
}