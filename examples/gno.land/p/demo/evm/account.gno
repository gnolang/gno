package evm

import (
	"encoding/hex"
	"errors"
	"std"

	"gno.land/p/demo/evm/keccak256"
	"gno.land/p/demo/uint256"
)

type Accounts map[uint64]*Account
type accountType uint

const (
	CallerAccount accountType = iota
	ContractAccount
)

func (x accountType) String() string {
	switch x {
	case CallerAccount:
		return "CallerAccount"
	case ContractAccount:
		return "ContractAccount"
	}
	return "<none>"
}

type Account struct {
	AccountType accountType
	EvmAddress  uint64
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

func (c *Account) executeCallHex(env *Env, valueToSend *uint256.Uint, f string, code []byte, args ...*uint256.Uint) ([]byte, error) {
	if valueToSend.Cmp(c.Balance) == 1 {
		return []byte{}, notEnoughWei
	}
	st := newStateFromEnv(env)
	buf := keccak256.Keccak256([]byte(f))[0:4]
	for _, v := range args {
		b := v.Bytes32()
		buf = append(buf, b[:]...)
	}
	st.setCallData(buf)
	st.setGasLimit(c.Gas)
	st.setCallerAddress(c.EvmAddress)
	st.setCode(code)
	err := st.executeByteCode()
	if err != nil && err != revertError {
		return []byte{}, err
	}

	c.Gas -= st.gasConsumed
	c.Balance.Sub(c.Balance, valueToSend)
	if err != revertError {
		st.loadStagedChangesIntoEnv(env)
	}
	return st.getResult(), nil
}

func (c *Account) Call(env *Env, valueToSend *uint256.Uint, code string, f string, args ...*uint256.Uint) ([]byte, error) {
	byteCode, err := hex.DecodeString(code)
	if err != nil {
		return []byte{}, err
	}
	return c.executeCallHex(env, valueToSend, f, byteCode, args...)
}

func (c *Account) CallContract(env *Env, contractAddress uint64, valueToSend *uint256.Uint, f string, args ...*uint256.Uint) ([]byte, error) {
	contract := env.Accounts[contractAddress]
	if contract.AccountType != ContractAccount {
		return []byte{}, errors.New("invalid contract address")
	}
	return c.executeCallHex(env, valueToSend, f, contract.Code, args...)
}

func (c *Accounts) PushCaller(env *Env, gnoAddress std.Address) (uint64, error) {
	for env.NextAccountId = env.NextAccountId + 1; (*c)[env.NextAccountId] != nil; env.NextAccountId = env.NextAccountId + 1 {
	}

	(*c)[env.NextAccountId] = &Account{
		AccountType: CallerAccount,
		EvmAddress:  env.NextAccountId,
		GnoAddress:  gnoAddress,
		Balance:     uint256.Zero(),
		Gas:         0,
		Code:        []byte{},
	}
	return env.NextAccountId, nil
}
