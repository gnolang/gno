package evm

import (
	"encoding/hex"
	"std"

	"gno.land/p/demo/evm/keccak256"
	"gno.land/p/demo/uint256"
)

type Caller struct {
	EVMAddress *uint256.Uint
	GnoAddress std.Address
	Balance    *uint256.Uint
	Gas        uint
}

func mustHex(str string) []byte {
	x, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return x
}

func (c *Caller) executeCallHex(env *Env, gasAmount uint, f string, code []byte, args ...*uint256.Uint) ([]byte, error) {
	if gasAmount > c.Gas {
		return []byte{}, notEnoughGas
	}
	st := newState()
	buf := keccak256.Keccak256([]byte(f))[0:4]
	for _, v := range args {
		b := v.Bytes32()
		buf = append(buf, b[:]...)
	}
	st.setCallData(buf)
	st.setEnv(env)
	st.setGasLimit(gasAmount)
	st.setCaller(c)
	err := st.executeByteCode(code)
	if err != nil {
		return []byte{}, err
	}
	return st.getResult(), nil
}

func (c *Caller) RawContractCall(env *Env, gasAmount uint, f string, raw string, args ...*uint256.Uint) ([]byte, error) {
	st := newState()
	err := st.executeByteCode(mustHex(raw))
	if err != nil {
		return []byte{}, err
	}
	return c.executeCallHex(env, gasAmount, f, st.getResult(), args...)
}

func (c *Caller) Call(env *Env, gasAmount uint, f string, code string, args ...*uint256.Uint) ([]byte, error) {
	x, err := hex.DecodeString(code)
	if err != nil {
		return []byte{}, err
	}
	return c.executeCallHex(env, gasAmount, f, x, args...)
}
