package evm

import (
	"errors"
	"strconv"

	"gno.land/p/demo/evm/keccak256"
	"gno.land/p/demo/uint256"
)

func opPush0(s *State) error {
	s.stack.Push(uint256.Zero())
	return nil
}

func opPush(n int) func(*State) error {
	return func(s *State) error {
		s.code.Increment()
		if s.code.Ill() {
			return errors.New("can't PUSH" + strconv.Itoa(n) + ": " + outOfBounds.Error())
		}
		stackData, err := s.code.Read(n)
		if err != nil {
			return errors.New("can't PUSH" + strconv.Itoa(n) + ": " + err.Error())
		}
		s.stack.Push(stackData)
		return nil
	}
}

func opSwap(n int) func(*State) error {
	return func(s *State) error {
		if s.stack.Len() <= n {
			return errors.New("can't SWAP" + strconv.Itoa(n) + ": stack too small")
		}
		temp := make([]*uint256.Uint, n+1)
		for i := 0; i <= n; i++ {
			val, err := s.stack.Pop()
			if err != nil {
				for j := i - 1; j >= 0; j-- {
					s.stack.Push(temp[j])
				}
				return errors.New("can't SWAP" + strconv.Itoa(n) + err.Error())
			}
			temp[i] = val
		}
		temp[0], temp[n] = temp[n], temp[0]
		for i := n; i >= 0; i-- {
			s.stack.Push(temp[i])
		}
		return nil
	}
}

func opDup(n int) func(*State) error {
	return func(s *State) error {
		if s.stack.Len() < n {
			return errors.New("can't DUP" + strconv.Itoa(n) + ": stack too small")
		}
		val, err := s.stack.Peek(n - 1)
		if err != nil {
			return errors.New("can't DUP" + strconv.Itoa(n) + ": " + err.Error())
		}

		s.stack.Push(val)
		return nil
	}
}

func opPop(s *State) error {
	_, err := s.stack.Pop()
	return err
}

func opJump(s *State) error {
	dest, err := s.stack.Pop()
	if err != nil {
		return err
	}
	inc := int(dest.Uint64())
	opDest := OpCode(s.code.At(inc))
	if opDest != JUMPDEST {
		return errors.New("can't JUMP: JUMP needs to land on a JUMPDEST, but here, it's a 0x" + strconv.FormatUint(uint64(opDest), 16) + ".")
	}
	s.code.rip = inc// + 1
	return nil
}

func opNop(s *State) error {
	return nil
}

func opJumpi(s *State) error {
	dest, err := s.stack.Pop()
	if err != nil {
		return err
	}
	doJump, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if doJump.IsZero() {
		return nil
	}
	inc := int(dest.Uint64())
	opDest := OpCode(s.code.At(inc))
	if opDest != JUMPDEST {
		return errors.New("can't JUMP: JUMP needs to land on a JUMPDEST, but here, it's a 0x" + strconv.FormatUint(uint64(opDest), 16) + ".")
	}
	s.code.rip = inc// + 1
	return nil
}

func opIszero(s *State) error {
	num, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if num.IsZero() {
		s.stack.Push(uint256.NewUint(1))
	} else {
		s.stack.Push(uint256.Zero())
	}
	return nil
}

func opKeccak256(s *State) error {
	off, err := s.stack.Pop()
	if err != nil {
		return err
	}
	size, err := s.stack.Pop()
	if err != nil {
		return err
	}
	off64 := off.Uint64()
	if off64 > uint64(len(s.memory)) {
		s.stack.Push(uint256.Zero())
		return nil
	}
	
	minWordSize := uint256.Zero().Div(uint256.Zero().Add(size, uint256.NewUint(31)), uint256.NewUint(32))
	s.gasConsumed += uint(6 * minWordSize.Uint64()) // TODO: runtime gas with mem extension (idk how to expand)

	slice := s.memory[off64:off64+size.Uint64()]
	resBytes := keccak256.Keccak256(slice)
	resUint := uint256.Zero().SetBytes(resBytes)
	s.stack.Push(resUint)
	return nil
}

func opPc(s *State) error {
	s.stack.Push(uint256.NewUint(uint64(s.code.rip)))
	return nil
}

func opCodeSize(s *State) error {
	s.stack.Push(uint256.NewUint(uint64(s.code.length)))
	return nil
}
