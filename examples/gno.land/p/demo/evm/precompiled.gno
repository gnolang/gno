package evm

var precompiled = map[uint64]func(s *State) ([]byte, error){
	(2): func(s *State) ([]byte, error) {
		return []byte{}, nil
	},
}
