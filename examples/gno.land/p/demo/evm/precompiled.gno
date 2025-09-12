package evm

import "crypto/sha256"

type precompiledAddress uint64

const (
	pEcRecover = iota + 1
	pSha256
	pRipemd160
	pIdentity
	pModexp
	pEcAdd
	pEcMul
	pEcPairing
	pBlake2f
	pPointEvaluation
)

var precompiled = map[precompiledAddress]func(s *state) error{
	(pEcRecover): func(s *state) error {
		return nil
	},
	(pSha256): func(s *state) error {
		res := sha256.Sum256(s.callData)
		s.result = res[:]
		s.gasConsumed = uint(60 + 12*(len(s.callData)+31)/32)
		return nil
	},
	(pRipemd160): func(s *state) error {
		return nil
	},
	(pIdentity): func(s *state) error {
		s.result = s.callData
		s.gasConsumed = uint(15 + 3*(len(s.callData)+31)/32)
		return nil
	},
	(pModexp): func(s *state) error {
		return nil
	},
	(pEcAdd): func(s *state) error {
		return nil
	},
	(pEcMul): func(s *state) error {
		return nil
	},
	(pEcPairing): func(s *state) error {
		return nil
	},
	(pBlake2f): func(s *state) error {
		return nil
	},
	(pPointEvaluation): func(s *state) error {
		return nil
	},
}
