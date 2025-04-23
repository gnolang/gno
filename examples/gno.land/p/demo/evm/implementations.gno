package evm

import (
	"errors"
	"strconv"

	"gno.land/p/demo/evm/uint256"
)

var outOfBounds = errors.New("out of bounds")

func opNotImplemented(s *state) error {
	return errors.New("op code " + strconv.FormatInt(int64(s.code.Current()), 16) + ": not implemented")
}

func opPush0(s *state) error {
	s.stack.Push(uint256.New())
	s.code.Increment()
	return nil
}

func opPush(n int) func(*state) error {
	return func(s *state) error {
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

func opSwap(n int) func(*state) error {
	return func(s *state) error {
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

		s.code.Increment()
		return nil
	}
}

func opPop(s *state) error {
	_, err := s.stack.Pop()
	return err
}

func opJump(s *state) error {
	dest, err := s.stack.Pop()
	if err != nil {
		return err
	}
	inc := int(dest.To64())
	opDest := OpCode(s.code.At(inc))
	if opDest != JUMPDEST {
		return errors.New("can't JUMP: JUMP needs to land on a JUMPDEST, but here, it's a 0x" + strconv.FormatUint(uint64(opDest), 16) + ".")
	}
	s.code.rip = inc + 1
	return nil
}

func opNop(s *state) error {
	s.code.Increment()
	return nil
}

func opJumpi(s *state) error {
	dest, err := s.stack.Pop()
	if err != nil {
		return err
	}
	doJump, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if doJump.IsZero() {
		s.code.Increment()
		return nil
	}
	inc := int(dest.To64())
	opDest := OpCode(s.code.At(inc))
	if opDest != JUMPDEST {
		return errors.New("can't JUMP: JUMP needs to land on a JUMPDEST, but here, it's a 0x" + strconv.FormatUint(uint64(opDest), 16) + ".")
	}
	s.code.rip = inc + 1
	return nil
}

func opIszero(s *state) error {
	num, err := s.stack.Pop()
	if err != nil {
		return err
	}
	if num.IsZero() {
		s.stack.Push(uint256.NewFrom64s(1))
	} else {
		s.stack.Push(uint256.New())
	}
	s.code.Increment()
	return nil
}

func opMload(s *state) error {
	off, err := s.stack.Pop()
	if err != nil {
		return err
	}
	off64 := off.To64()
	if off64 + 32 > uint64(len(s.memory)) {
		s.stack.Push(uint256.New())
		s.code.Increment()
		return nil
	}
	read := s.memory[off64:off64+32]
	if len(read) < 32 {
		read = append(read, make([]byte, 32-len(read))...)
	}
	s.stack.Push(uint256.NewFromBytes(read))
	s.code.Increment()
	return nil
}

func opMstore(s *state) error {
	off, err := s.stack.Pop()
	if err != nil {
		return err
	}
	off64 := off.To64()
	val, err := s.stack.Pop()
	if err != nil {
		return err
	}
	bytes := val.ToBytes()
	memLen := uint64(len(s.memory))
	if memLen == 0 {
		s.memory = append(make([]byte, off64), bytes[:]...)
		s.code.Increment()
		return nil
	}
	if off64 + 32 >= memLen {
		s.memory = append(s.memory, make([]byte, off64 + 32 - memLen)...)
	}
	copy(s.memory[off64:off64+32], bytes[:])
	s.code.Increment()
	return nil
}

func opMstore8(s *state) error {
	off, err := s.stack.Pop()
	if err != nil {
		return err
	}
	off64 := off.To64()
	val, err := s.stack.Pop()
	if err != nil {
		return err
	}
	b := byte(val.To64() & 0xff)
	memLen := uint64(len(s.memory))
	if memLen == 0 && off64 != 0 {
		s.memory = append(make([]byte, off64), b)
		s.code.Increment()
		return nil
	}
	if off64 >= memLen {
		s.memory = append(s.memory, make([]byte, off64)...)
	}
	s.memory = append(s.memory, b)
	s.code.Increment()
	return nil
}
