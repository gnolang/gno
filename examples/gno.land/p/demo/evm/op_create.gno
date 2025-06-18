package evm

import "gno.land/p/demo/uint256"

func opCreate(s *State) error {
	// pop arguments from stack
	value, err := s.stack.Pop()
	if err != nil {
		return err
	}
	offset, err := s.stack.Pop()
	if err != nil {
		return err
	}
	size, err := s.stack.Pop()
	if err != nil {
		return err
	}

	// check balance
	if s.tmp.accounts[s.callerAddress].Balance.Cmp(value) == -1 {
		s.stack.Push(uint256.Zero())
		return nil
	}

	initCode := getByteCopyFromMemory(s.memory, offset, size)

	for s.env.NextAccountId = s.env.NextAccountId + 1; s.tmp.accounts[s.env.NextAccountId] != nil; s.env.NextAccountId = s.env.NextAccountId + 1 {
	}

	// check if address already exists
	if acc := s.tmp.accounts[s.env.NextAccountId]; acc != nil && (len(acc.Code) > 0 || acc.Balance.Sign() > 0 || acc.Nonce > 0) {
		s.stack.Push(uint256.Zero()) // address in use
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
		s.stack.Push(uint256.Zero())
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

	s.stack.Push(uint256.NewUint(s.env.NextAccountId))
	s.tmp.accounts[s.callerAddress].Balance.Sub(s.tmp.accounts[s.callerAddress].Balance, value)
	s.tmp.warmStorage = append(s.tmp.warmStorage, *uint256.NewUint(s.env.NextAccountId))
	s.gasConsumed += uint(uint64(len(runtimeCode)*200) + 2*((size.Uint64()+31)/32))
	return nil
}
