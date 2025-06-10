package evm

import "crypto/sha256"

type precompiledAddress uint64
const (
	pEcRecover = iota+1
	pSha256
)

var precompiled = map[precompiledAddress]func(s *State) error {
	(pSha256): func(s *State) error {
		res := sha256.Sum256(s.callData)
		s.result = res[:]
		return nil
	},
}
