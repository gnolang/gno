package evm

import (
	"encoding/hex"
	"errors"
	"std"

	"gno.land/p/demo/evm/keccak256"
	"gno.land/p/demo/uint256"
)

type Accounts map[uint256.Uint]*Account
type accountType int

const (
	CallerAccount accountType = iota
	ContractAccount
)

type Account struct {
	AccountType accountType
	EvmAddress  *uint256.Uint
	GnoAddress  std.Address
	Balance     *uint256.Uint
	Gas         uint
	Code        []byte
}

func mustHex(str string) []byte {
	x, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return x
}

func (c *Account) executeCallHex(env *Env, gasAmount uint, f string, code []byte, args ...*uint256.Uint) ([]byte, error) {
	if gasAmount > c.Gas {
		return []byte{}, notEnoughGas
	}
	st := newStateFromEnv(env)
	buf := keccak256.Keccak256([]byte(f))[0:4]
	for _, v := range args {
		b := v.Bytes32()
		buf = append(buf, b[:]...)
	}
	st.setCallData(buf)
	st.setGasLimit(gasAmount)
	st.setCallerAddress(c.EvmAddress)
	st.setCode(code)
	err := st.executeByteCode()
	if err != nil {
		return []byte{}, err
	}

	if st.callerAddress != nil && st.tmp.accounts[*st.callerAddress] != nil {
		st.tmp.accounts[*st.callerAddress].Gas -= st.gasConsumed
	}
	st.loadStagedChangesIntoEnv(env)
	return st.getResult(), nil
}

func (c *Account) RawContractCall(env *Env, gasAmount uint, f string, raw string, args ...*uint256.Uint) ([]byte, error) {
	st := newStateFromEnv(env)
	st.setCode(mustHex(raw))
	err := st.executeByteCode()
	if err != nil {
		return []byte{}, err
	}
	st.loadStagedChangesIntoEnv(env)
	return c.executeCallHex(env, gasAmount, f, st.getResult(), args...)
}

func (c *Account) Call(env *Env, gasAmount uint, f string, code string, args ...*uint256.Uint) ([]byte, error) {
	x, err := hex.DecodeString(code)
	if err != nil {
		return []byte{}, err
	}
	return c.executeCallHex(env, gasAmount, f, x, args...)
}

func (c *Accounts) PushSmartContract(env *Env, code []byte) (*uint256.Uint, error) {
	if c == nil {
		return nil, errors.New("Accounts are nil")
	}
	for (*c)[*env.NextAccountId.Add(&env.NextAccountId, uint256.One())] != nil {
	}

	(*c)[env.NextAccountId] = &Account{
		AccountType: ContractAccount,
		EvmAddress:  env.NextAccountId.Clone(),
		GnoAddress:  "",
		Balance:     uint256.Zero(),
		Gas:         0,
		Code:        code,
	}
	return env.NextAccountId.Clone(), nil
}

func (c *Accounts) PushCaller(env *Env, gnoAddress std.Address) (*uint256.Uint, error) {
	if c == nil {
		return nil, errors.New("Accounts are nil")
	}
	for (*c)[*env.NextAccountId.Add(&env.NextAccountId, uint256.One())] != nil {
	}

	(*c)[env.NextAccountId] = &Account{
		AccountType: CallerAccount,
		EvmAddress:  env.NextAccountId.Clone(),
		GnoAddress:  gnoAddress,
		Balance:     uint256.Zero(),
		Gas:         0,
		Code:        []byte{},
	}
	return env.NextAccountId.Clone(), nil
}
