package evm

import (
	"math"

	"gno.land/p/demo/uint256"
)

func opCall(s *State) error {
	gas, err := s.stack.Pop()
	if err != nil {
		return err
	}
	address, err := s.stack.Pop()
	if err != nil {
		return err
	}

	value, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if s.caller.Balance.Cmp(value) < 0 {
		s.stack.Push(uint256.Zero())
		return nil
	}

	argsOffset, err := s.stack.Pop()
	if err != nil {
		return err
	}
	argsSize, err := s.stack.Pop()
	if err != nil {
		return err
	}
	retOffset, err := s.stack.Pop()
	if err != nil {
		return err
	}
	retSize, err := s.stack.Pop()
	if err != nil {
		return err
	}

	memoryCost := uint64(0)
	tmpState := s.clone()

	callData := make([]byte, 0)
	if uint(gas.Uint64()) > tmpState.caller.Gas {
		return notEnoughGas
	}
	if int(argsOffset.Uint64()) >= len(s.memory) {
		callData = make([]byte, int(argsSize.Uint64()))
	} else if int(argsOffset.Uint64()+argsSize.Uint64()) > len(s.memory) {
		callData = make([]byte, int(argsSize.Uint64()))
		copy(callData, s.memory[argsOffset.Uint64():])
	} else {
		callData = tmpState.memory[argsOffset.Uint64() : argsOffset.Uint64()+argsSize.Uint64()]
	}
	tmpState.setCallData(callData)
	tmpState.setGasLimit(uint(gas.Uint64()))
	tmpState.setCaller(tmpState.caller)
	contracts := tmpState.tmp.contracts
	if address.Uint64() > uint64(len(contracts)) {
		s.stack.Push(uint256.Zero())
		return nil
	}
	contractCode := contracts[address.Uint64()].Code

	err = tmpState.executeByteCode(contractCode)
	if err != nil {
		s.stack.Push(uint256.Zero())
		return nil
	}
	s.stack.Push(uint256.One())
	memLen := uint64(len(s.memory))
	farest := retOffset.Uint64() + retSize.Uint64()
	if farest > memLen {
		extend := farest - memLen
		s.memory = append(s.memory, make([]byte, extend)...)

		memorySizeWord := (extend + 31) / 32
		memoryCost = uint64(math.Pow(float64(memorySizeWord), 2))/512 + (3 * memorySizeWord)
	}
	res := tmpState.getResult()
	copy(s.memory[retOffset.Uint64():retOffset.Uint64()+retSize.Uint64()], res)
	s.nextCallState = tmpState

	s.gasConsumed += uint(memoryCost) + tmpState.gasConsumed
	s.tmp.storage = tmpState.tmp.storage
	s.tmp.logs = tmpState.tmp.logs
	s.tmp.accounts = tmpState.tmp.accounts
	s.tmp.contracts = tmpState.tmp.contracts
	s.tmp.warmStorage = tmpState.tmp.warmStorage
	return nil
}

func opReturnDataSize(s *State) error {
	s.stack.Push(uint256.NewUint(uint64(len(s.result))))
	return nil
}
