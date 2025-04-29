package evm

import (
	"errors"
	"std"

	"gno.land/p/demo/uint256"
)

type Caller struct {
	EVMAddress *uint256.Uint
	GnoAddress std.Address
	Gas uint
}

var (
	notEnoughGas = errors.New("the user does not have that much gas")
)

func (c *Caller) Call(env *Env, gasAmount uint, code []byte) error {
	if gasAmount > c.Gas {
		return notEnoughGas
	}
	st := newState()
	st.openEnv(env)
	err := st.executeByteCodeWithGas(code, int(gasAmount))
	if err != nil {
		return err
	}
	return nil
}