package evm

import "gno.land/p/demo/uint256"

// Retrieve data stored in memory, at offset.
// > IN: `[offset]`
// > OUT: `[result]`
func opMload(s *state) error {
	off, err := s.stack.Pop()
	if err != nil {
		return err
	}
	off64 := off.Uint64()
	if off64+32 > uint64(len(s.memory)) {
		s.stack.Push(uint256.Zero())
		return nil
	}

	read := s.memory[off64 : off64+32]
	if len(read) < 32 {
		read = append(read, make([]byte, 32-len(read))...)
	}
	s.stack.Push(uint256.Zero().SetBytes(read))
	return nil
}

// Write data in memory at offset.
// > IN: `[offset, data]`
func opMstore(s *state) error {
	off, err := s.stack.Pop()
	if err != nil {
		return err
	}
	off64 := off.Uint64()
	val, err := s.stack.Pop()
	if err != nil {
		return err
	}
	bytes := val.Bytes32()
	memLen := uint64(len(s.memory))
	if memLen == 0 {
		s.memory = append(make([]byte, off64), bytes[:]...)
		return nil
	}
	if off64+32 >= memLen {
		s.memory = append(s.memory, make([]byte, off64+32-memLen)...)
	}
	copy(s.memory[off64:off64+32], bytes[:])
	return nil
}

// Write first byte of data in memory at offset.
// > IN: `[offset, data]`
func opMstore8(s *state) error {
	off, err := s.stack.Pop()
	if err != nil {
		return err
	}
	off64 := off.Uint64()
	val, err := s.stack.Pop()
	if err != nil {
		return err
	}
	b := byte(val.Uint64() & 0xff)
	memLen := uint64(len(s.memory))
	if memLen == 0 && off64 != 0 {
		s.memory = append(make([]byte, off64), b)
		return nil
	}
	if off64 + 1 > memLen {
		s.memory = append(s.memory, make([]byte, off64 + 1 - memLen)...)
	}
	s.memory[off64] = b
	return nil
}

// Get the memory size.
// > OUT: `[size]`
func opMsize(s *state) error {
	s.stack.Push(uint256.NewUint(uint64(len(s.memory))))
	return nil
}
