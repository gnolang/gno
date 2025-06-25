package evm

import (
	"encoding/hex"
	"errors"
	"std"
	"strconv"

	"gno.land/p/demo/evm/keccak256"
	"gno.land/p/demo/uint256"
)

type Accounts map[uint64]*Account
type AccountType uint

const (
	CallerAccount AccountType = iota
	ContractAccount
)

func (x AccountType) String() string {
	switch x {
	case CallerAccount:
		return "Caller Account"
	case ContractAccount:
		return "Contract Account"
	default:
		return "Unknown Account type"
	}
}

type Account struct {
	AccountType AccountType
	EvmAddress  uint64
	GnoAddress  std.Address
	Balance     *uint256.Uint
	Nonce       uint64
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

func (c *Account) executeCallHex(env *Env, f string, code []byte, args ...*uint256.Uint) ([]byte, error) {
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
	if valueToSend.Gt(c.Balance) {
		return []byte{}, notEnoughWei
	}
	c.Balance.Sub(c.Balance, valueToSend)
	return c.executeCallHex(env, f, byteCode, args...)
}

func (c *Account) CallContract(env *Env, contractAddress uint64, valueToSend *uint256.Uint, f string, args ...*uint256.Uint) ([]byte, error) {
	contract, ok := env.Accounts[contractAddress]
	if !ok || contract.AccountType != ContractAccount {
		return []byte{}, errors.New("invalid contract address")
	}
	if valueToSend.Gt(c.Balance) {
		return []byte{}, notEnoughWei
	}
	c.Balance.Sub(c.Balance, valueToSend)
	c.Balance.Add(contract.Balance, valueToSend)
	return c.executeCallHex(env, f, contract.Code, args...)
}

func (c *Account) CreateContract(env *Env, code string, valueToSend *uint256.Uint) ([]byte, error) {
	byteCode, err := hex.DecodeString(code)
	if err != nil {
		return []byte{}, err
	}

	creationCode := []byte{}

	lenByteCodeUint256 := uint256.NewUint(uint64(len(byteCode)))
	creationCode = append(creationCode,
		byte(PUSH0)+byte(lenByteCodeUint256.ByteLen()),
	)
	creationCode = append(creationCode,
		lenByteCodeUint256.Bytes()...,
	)
	codeOffsetUint256 := uint256.NewUint(uint64(1 + lenByteCodeUint256.ByteLen() + 33 + 2 + 1 + lenByteCodeUint256.ByteLen() + 1 + 1 + valueToSend.ByteLen() + 2)).Bytes32()
	creationCode = append(creationCode,
		byte(PUSH32),
	)
	creationCode = append(creationCode,
		codeOffsetUint256[:]...,
	)
	creationCode = append(creationCode,
		byte(PUSH0),
	)
	creationCode = append(creationCode, byte(CODECOPY))

	creationCode = append(creationCode,
		byte(PUSH0)+byte(lenByteCodeUint256.ByteLen()),
	)
	creationCode = append(creationCode,
		lenByteCodeUint256.Bytes()...,
	)
	creationCode = append(creationCode,
		byte(PUSH0),
	)
	creationCode = append(creationCode,
		byte(PUSH0)+byte(valueToSend.ByteLen()),
	)
	creationCode = append(creationCode,
		valueToSend.Bytes()...,
	)
	creationCode = append(creationCode, byte(CREATE))
	creationCode = append(creationCode, byte(STOP))

	creationCode = append(creationCode, byteCode...)

	return c.executeCallHex(env, "", creationCode)
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

func (c *Account) String() string {
	res := "AccountType: " + c.AccountType.String()
	res += ", EvmAddress: " + uint256.NewUint(c.EvmAddress).Hex()
	res += ", GnoAddress: " + c.GnoAddress.String()
	res += ", Balance: " + c.Balance.String()
	res += ", Nonce: " + strconv.FormatUint(c.Nonce, 10)
	res += ", Gas: " + strconv.FormatUint(uint64(c.Gas), 10)

	return res
}

func (c *Account) Eq(a Account) bool {
	return c.String() == a.String()
}
