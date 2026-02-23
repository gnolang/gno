package evm

import (
	"errors"

	"gno.land/p/demo/evm/keccak256"
	"gno.land/p/onbloc/uint256"
)

func opCreate(s *state) error {
	if s.context == staticContext {
		return errors.New("can't create from staticCall")
	}

	// pop arguments from stack
	value, err := s.stack.pop()
	if err != nil {
		return err
	}
	offset, err := s.stack.pop()
	if err != nil {
		return err
	}
	size, err := s.stack.pop()
	if err != nil {
		return err
	}

	// check balance
	if s.tmp.accounts[s.callerAddress].Balance.Cmp(value) == -1 {
		s.stack.push(uint256.Zero())
		return nil
	}

	initCode := getByteCopyFromMemory(s.memory, offset, size)

	for s.env.NextAccountId = s.env.NextAccountId + 1; s.tmp.accounts[s.env.NextAccountId] != nil; s.env.NextAccountId = s.env.NextAccountId + 1 {
	}

	// create substate for contract execution
	tmpState := s.clone()
	tmpState.setCallData(nil)
	tmpState.setCode(initCode)

	err = tmpState.executeByteCode()
	s.gasConsumed += tmpState.gasConsumed

	// creation failed
	if err != nil {
		s.stack.push(uint256.Zero())
		return nil
	}

	runtimeCode := tmpState.getResult()

	s.tmp.accounts[s.env.NextAccountId] = &Account{
		AccountType: ContractAccount,
		EvmAddress:  s.env.NextAccountId,
		GnoAddress:  "",
		Balance:     value.Clone(),
		Gas:         0,
		Code:        runtimeCode,
	}

	s.stack.push(uint256.NewUint(s.env.NextAccountId))
	s.tmp.accounts[s.callerAddress].Balance.Sub(s.tmp.accounts[s.callerAddress].Balance, value)
	s.tmp.warmAccountAddress = append(s.tmp.warmAccountAddress, *uint256.NewUint(s.env.NextAccountId))
	minWordSize := (size.Uint64() + 31) / 32

	s.gasConsumed += uint(uint64(len(runtimeCode)*200) + 2*minWordSize)
	return nil
}

func opCreate2(s *state) error {
	if s.context == staticContext {
		return errors.New("can't create2 from staticCall")
	}

	// pop arguments from stack
	value, err := s.stack.pop()
	if err != nil {
		return err
	}
	offset, err := s.stack.pop()
	if err != nil {
		return err
	}
	size, err := s.stack.pop()
	if err != nil {
		return err
	}
	salt, err := s.stack.pop()
	if err != nil {
		return err
	}

	// check balance
	if s.tmp.accounts[s.callerAddress].Balance.Cmp(value) == -1 {
		s.stack.push(uint256.Zero())
		return nil
	}

	initCode := getByteCopyFromMemory(s.memory, offset, size)

	// compute new adress
	initCodeKeccak256 := uint256.Zero().SetBytes(keccak256.Keccak256(initCode))
	rawNewAddress256 := initCodeKeccak256.Add(initCodeKeccak256, uint256.NewUint(0xff))
	rawNewAddress256 = rawNewAddress256.Add(rawNewAddress256, uint256.NewUint(s.callerAddress))
	rawNewAddress256 = rawNewAddress256.Add(rawNewAddress256, salt)

	rawNewAddressBytes32 := rawNewAddress256.Bytes32()

	newAccountAddr := keccak256.Keccak256(rawNewAddressBytes32[:])[12:]
	newAccountAddr256 := uint256.Zero().SetBytes(newAccountAddr).Uint64()

	// check if address already exists
	if acc := s.tmp.accounts[newAccountAddr256]; acc != nil && (len(acc.Code) > 0 || acc.Balance.Sign() > 0 || acc.Nonce > 0) {
		s.stack.push(uint256.Zero()) // address in use
		return nil
	}

	// create substate for contract execution
	tmpState := s.clone()
	tmpState.setCallData(nil)
	tmpState.setCode(initCode)

	err = tmpState.executeByteCode()
	s.gasConsumed += tmpState.gasConsumed

	// creation failed
	if err != nil {
		s.stack.push(uint256.Zero())
		return nil
	}

	runtimeCode := tmpState.getResult()

	s.tmp.accounts[newAccountAddr256] = &Account{
		AccountType: ContractAccount,
		EvmAddress:  newAccountAddr256,
		GnoAddress:  "",
		Balance:     value.Clone(),
		Gas:         0,
		Code:        runtimeCode,
	}

	s.stack.push(uint256.NewUint(newAccountAddr256))
	s.tmp.accounts[s.callerAddress].Balance.Sub(s.tmp.accounts[s.callerAddress].Balance, value)
	s.tmp.warmAccountAddress = append(s.tmp.warmAccountAddress, *uint256.NewUint(newAccountAddr256))
	minWordSize := (size.Uint64() + 31) / 32

	s.gasConsumed += uint(uint64(len(runtimeCode)*200) + 2*minWordSize + 6*minWordSize)
	return nil
}
