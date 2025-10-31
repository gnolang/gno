package evm

import (
	"encoding/hex"
	"errors"
	"std"
	"strconv"

	"gno.land/p/demo/evm/keccak256"
	"gno.land/p/onbloc/uint256"
)

// Accounts is a map of EVM addresses to Account pointers
type Accounts map[uint64]*Account

// AccountType represents the type of an account: caller or contract
type AccountType uint

const (
	CallerAccount   AccountType = iota // Externally owned account
	ContractAccount                    // Smart contract account
)

// AccountType as string
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

// Account represents an EVM account
type Account struct {
	AccountType AccountType              // Type of the account: Caller or Contract
	EvmAddress  uint64                   // Unique EVM address (uint64)
	GnoAddress  std.Address              // Corresponding Gno address
	Balance     *uint256.Uint            // Account balance in Wei
	Nonce       uint64                   // Nonce for transaction ordering
	Gas         uint                     // Available gas
	Storage     map[uint64]*uint256.Uint // Contract key-value storage
	Code        []byte                   // Smart contract bytecode (if any)
}

func mustHex(str string) []byte {
	x, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return x
}

// Clone creates a deep copy of the Accounts map, cloning each Account within it.
func (c *Accounts) Clone() *Accounts {
	accs := make(Accounts)
	if len(*c) == 0 {
		return &accs
	}
	for k, v := range *c {
		accs[k] = v.Clone()
	}
	return &accs
}

func (c *Account) executeCallHex(env *Env, gasLimit int, f string, code []byte, args ...*uint256.Uint) ([]byte, uint, error) {
	st := newStateFromEnv(env)
	buf := keccak256.Keccak256([]byte(f))[0:4]
	for _, v := range args {
		b := v.Bytes32()
		buf = append(buf, b[:]...)
	}
	st.setCallData(buf)
	st.setGasLimit(gasLimit)
	st.setCallerAddress(c.EvmAddress)
	st.setOriginAddress(c.EvmAddress)
	st.setCode(code)
	err := st.executeByteCode()
	if err != nil && err != revertError {
		return []byte{}, st.gasConsumed, err
	}

	if gasLimit != -1 {
		c.Gas -= st.gasConsumed
	}
	if err != revertError {
		st.loadStagedChangesIntoEnv(env)
	}
	return st.getResult(), st.gasConsumed, nil
}

// Call executes raw byte code (hex string) deducting the valueToSend from the caller's account
// Returns the output, gas consumed, and any error
func (c *Account) Call(env *Env, gasLimit int, valueToSend *uint256.Uint, code string, f string, args ...*uint256.Uint) ([]byte, uint, error) {
	byteCode, err := hex.DecodeString(code)
	if err != nil {
		return []byte{}, 0, err
	}
	if valueToSend.Gt(c.Balance) {
		return []byte{}, 0, notEnoughWei
	}
	c.Balance.Sub(c.Balance, valueToSend)
	return c.executeCallHex(env, gasLimit, f, byteCode, args...)
}

// CallContract call an already deployed contract (by address), transferring the valueToSend from caller's account to the contract account
// Returns the output, gas consumed, and any error
func (c *Account) CallContract(env *Env, gasLimit int, contractAddress uint64, valueToSend *uint256.Uint, f string, args ...*uint256.Uint) ([]byte, uint, error) {
	contract, ok := env.Accounts[contractAddress]
	if !ok || contract.AccountType != ContractAccount {
		return []byte{}, 0, errors.New("invalid contract address")
	}
	if valueToSend.Gt(c.Balance) {
		return []byte{}, 0, notEnoughWei
	}
	c.Balance.Sub(c.Balance, valueToSend)
	c.Balance.Add(contract.Balance, valueToSend)
	return c.executeCallHex(env, gasLimit, f, contract.Code, args...)
}

// CreateContract create a new contract from only it's creation code, transferring the valueToSend from caller's account to the newly created contract account
// Returns the output, gas consumed, and any error
func (c *Account) CreateContract(env *Env, gasLimit int, code string, valueToSend *uint256.Uint) (uint64, uint, error) {
	byteCode, err := hex.DecodeString(code)
	if err != nil {
		return 0, 0, err
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

	_, gasConsumed, err := c.executeCallHex(env, gasLimit, "", creationCode)

	if err != nil {
		return 0, gasConsumed, err
	}

	return 0, gasConsumed, nil
}

// PushCaller creates a new caller account with the given Gno address,
// assigns a new EVM address, and adds it to the Accounts map.
// Returns the assigned EVM address.
func (c *Accounts) PushCaller(env *Env, gnoAddress std.Address) (uint64, error) {
	for env.NextAccountId = env.NextAccountId + 1; (*c)[env.NextAccountId] != nil; env.NextAccountId = env.NextAccountId + 1 {
	}

	(*c)[env.NextAccountId] = &Account{
		AccountType: CallerAccount,
		EvmAddress:  env.NextAccountId,
		GnoAddress:  gnoAddress,
		Balance:     uint256.Zero(),
		Gas:         0,
		Storage:     map[uint64]*uint256.Uint{},
		Code:        []byte{},
	}
	return env.NextAccountId, nil
}

// String returns a string representation of the Account,
// including its type, address, balance, nonce, gas, and partial code bytes.
func (c *Account) String() string {
	res := "AccountType: " + c.AccountType.String()
	res += ", EvmAddress: " + uint256.NewUint(c.EvmAddress).Hex()
	res += ", GnoAddress: " + c.GnoAddress.String()
	res += ", Balance: " + c.Balance.String()
	res += ", Nonce: " + strconv.FormatUint(c.Nonce, 10)
	res += ", Gas: " + strconv.FormatUint(uint64(c.Gas), 10)
	res += ", Code: len = " + strconv.Itoa(len(c.Code)) + " : [ "

	for i := 0; i < 10 && i < len(c.Code); i += 1 {
		res += string(c.Code[i]) + ", "
	}
	if len(c.Code) >= 10 {
		res += "... "
	}
	res += "]"

	return res
}

func (c *Account) Eq(a Account) bool {
	return c.String() == a.String()
}

// Clone creates a deep copy of the Account, including storage, balance, and code
func (c *Account) Clone() *Account {
	storageCopy := make(map[uint64]*uint256.Uint)
	if len(c.Storage) > 0 {
		for k, v := range c.Storage {
			storageCopy[k] = v.Clone()
		}
	}

	codeCopy := make([]byte, len(c.Code))
	copy(codeCopy, c.Code)

	return &Account{
		AccountType: c.AccountType,
		EvmAddress:  c.EvmAddress,
		GnoAddress:  c.GnoAddress,
		Balance:     c.Balance.Clone(),
		Nonce:       c.Nonce,
		Gas:         c.Gas,
		Storage:     storageCopy,
		Code:        codeCopy,
	}
}
