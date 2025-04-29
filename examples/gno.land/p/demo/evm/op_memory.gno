package evm

import (
	"gno.land/p/demo/uint256"
)

// Retrieve data stored in memory, at offset.
// > IN: `[offset]`
// > OUT: `[result]`
func opMload(s *State) error {
	off, err := s.stack.Pop()
	if err != nil {
		return err
	}
	off64 := off.Uint64()
	if off64+32 > uint64(len(s.memory)) {
		s.stack.Push(uint256.Zero())
		return nil
	}

	// TODO: add read-memory gas consuption
	read := s.memory[off64 : off64+32]
	if len(read) < 32 {
		read = append(read, make([]byte, 32-len(read))...)
	}
	s.stack.Push(uint256.Zero().SetBytes(read))
	return nil
}

// Write data in memory at offset.
// > IN: `[offset, data]`
func opMstore(s *State) error {
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

	// if memory is empty, add the bytes, expension is 32 bytes
	if memLen == 0 {
		s.memory = append(make([]byte, off64), bytes[:]...)
		s.gasConsumed += 32
		return nil
	}
	// if offset + 32 is greater than memory len, expend/consume the diff
	if off64+32 >= memLen {
		add := off64 + 32 - memLen
		s.memory = append(s.memory, make([]byte, add)...)
		s.gasConsumed += uint(add)
	}

	// copy in the right place
	copy(s.memory[off64:off64+32], bytes[:])
	return nil
}

// Write first byte of data in memory at offset.
// > IN: `[offset, data]`
func opMstore8(s *State) error {
	off, err := s.stack.Pop()
	if err != nil {
		return err
	}
	off64 := off.Uint64()
	val, err := s.stack.Pop()
	if err != nil {
		return err
	}

	// get the byte
	b := byte(val.Uint64() & 0xff)
	memLen := uint64(len(s.memory))

	// if no mem and offset is greater than zero, just do some zeroes + add the byte
	if memLen == 0 && off64 != 0 {
		s.memory = append(make([]byte, off64), b)
		s.gasConsumed += uint(off64) + 1
		return nil
	}
	if off64+1 > memLen {
		add := off64 + 1 - memLen
		s.memory = append(s.memory, make([]byte, add)...)
		s.gasConsumed += uint(add)
	}
	s.memory[off64] = b
	return nil
}

// Get the memory size.
// > OUT: `[size]`
func opMsize(s *State) error {
	s.stack.Push(uint256.NewUint(uint64(len(s.memory))))
	return nil
}

func opMcopy(s *State) error {
	dst, err := s.stack.Pop()
	if err != nil {
		return err
	}
	src, err := s.stack.Pop()
	if err != nil {
		return err
	}
	size, err := s.stack.Pop()
	if err != nil {
		return err
	}

	memlen256 := uint256.NewUint(uint64(len(s.memory)))
	farest := uint256.Zero().Max(uint256.Zero().Add(src, size), uint256.Zero().Add(dst, size))
	if farest.Cmp(memlen256) > 0 {
		extend := uint256.Zero().Sub(farest, memlen256).Uint64()
		s.memory = append(s.memory, make([]byte, extend)...)
		s.gasConsumed += uint(extend)
	}
	copy(s.memory[dst.Uint64():], s.memory[src.Uint64():src.Uint64()+size.Uint64()])
	return nil
}

func opTload(s *State) error {
	key, err := s.stack.Pop()
	if err != nil {
		return err
	}

	read, ok := s.transientStorage[*key]
	if !ok {
		s.stack.Push(uint256.Zero())
	}
	s.stack.Push(read)
	return nil
}

func opTstore(s *State) error {
	key, err := s.stack.Pop()
	if err != nil {
		return err
	}
	value, err := s.stack.Pop()
	if err != nil {
		return err
	}
	s.transientStorage[*key] = value
	return nil
}

func opSload(s *State) error {
	key, err := s.stack.Pop()
	if err != nil {
		return err
	}

	warm := false
	for _, e := range s.tmp.warmStorage {
		if e.Eq(key) {
			warm = true
			break
		}
	}
	if !warm {
		s.tmp.warmStorage = append(s.tmp.warmStorage, *key)
		s.gasConsumed += 2000
	}

	read, ok := s.tmp.storage[*key]
	if !ok {
		s.stack.Push(uint256.Zero())
	}
	s.stack.Push(read)
	return nil
}

func opSstore(s *State) error {
	key, err := s.stack.Pop()
	if err != nil {
		return err
	}
	value, err := s.stack.Pop()
	if err != nil {
		return err
	}

	warm := false
	for _, e := range s.tmp.warmStorage {
		if e.Eq(key) {
			warm = true
			break
		}
	}
	if !warm {
		s.tmp.warmStorage = append(s.tmp.warmStorage, *key)
		s.gasConsumed += 2000
	}

	s.tmp.storage[*key] = value
	return nil
}
