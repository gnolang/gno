package evm

import (
	"errors"

	"gno.land/p/demo/uint256"
)

func opSelfDestruct(s *State) error {
	addr, err := s.stack.Pop()
	if err != nil {
		return errors.New("can't self destruct: need account address")
	}
	if addr.Eq(uint256.NewUint(s.callerAddress)) {
		return errors.New("can't self destruct: you are giving your balance to yourself")
	}
	if s.tmp.accounts[addr.Uint64()] == nil && !s.tmp.accounts[s.callerAddress].Balance.IsZero() {
		s.gasConsumed += 25000
	}
	warm := false
	for _, e := range s.tmp.warmAccountAddress {
		if e.Eq(addr) {
			warm = true
			break
		}
	}
	if !warm {
		s.tmp.warmAccountAddress = append(s.tmp.warmAccountAddress, *addr)
		s.gasConsumed += 2600
	}
	if s.tmp.accounts[addr.Uint64()] != nil {
		s.tmp.accounts[addr.Uint64()].Balance.Add(s.tmp.accounts[s.callerAddress].Balance, s.tmp.accounts[addr.Uint64()].Balance)
	}
	delete(s.tmp.accounts, s.callerAddress)
	return nil
}
