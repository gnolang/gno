package core

import "github.com/gnolang/go-tendermint/messages/types"

type (
	broadcastProposalDelegate  func(*types.ProposalMessage)
	broadcastPrevoteDelegate   func(*types.PrevoteMessage)
	broadcastPrecommitDelegate func(*types.PrecommitMessage)
)

type mockBroadcast struct {
	broadcastProposalFn  broadcastProposalDelegate
	broadcastPrevoteFn   broadcastPrevoteDelegate
	broadcastPrecommitFn broadcastPrecommitDelegate
}

func (m *mockBroadcast) BroadcastProposal(message *types.ProposalMessage) {
	if m.broadcastProposalFn != nil {
		m.broadcastProposalFn(message)
	}
}

func (m *mockBroadcast) BroadcastPrevote(message *types.PrevoteMessage) {
	if m.broadcastPrevoteFn != nil {
		m.broadcastPrevoteFn(message)
	}
}

func (m *mockBroadcast) BroadcastPrecommit(message *types.PrecommitMessage) {
	if m.broadcastPrecommitFn != nil {
		m.broadcastPrecommitFn(message)
	}
}

type (
	idDelegate            func() []byte
	hashDelegate          func([]byte) []byte
	buildProposalDelegate func(uint64) []byte
)

type mockNode struct {
	idFn            idDelegate
	hashFn          hashDelegate
	buildProposalFn buildProposalDelegate
}

func (m *mockNode) ID() []byte {
	if m.idFn != nil {
		return m.idFn()
	}

	return nil
}

func (m *mockNode) Hash(proposal []byte) []byte {
	if m.hashFn != nil {
		return m.hashFn(proposal)
	}

	return nil
}

func (m *mockNode) BuildProposal(height uint64) []byte {
	if m.buildProposalFn != nil {
		return m.buildProposalFn(height)
	}

	return nil
}

type (
	signDelegate             func([]byte) []byte
	isValidSignatureDelegate func([]byte, []byte) bool
)

type mockSigner struct {
	signFn             signDelegate
	isValidSignatureFn isValidSignatureDelegate
}

func (m *mockSigner) Sign(data []byte) []byte {
	if m.signFn != nil {
		return m.signFn(data)
	}

	return nil
}

func (m *mockSigner) IsValidSignature(data, signature []byte) bool {
	if m.isValidSignatureFn != nil {
		return m.isValidSignatureFn(data, signature)
	}

	return false
}

type (
	isProposerDelegate func([]byte, uint64, uint64) bool
	isValidator        func([]byte) bool
)

type mockVerifier struct {
	isProposerFn  isProposerDelegate
	isValidatorFn isValidator
}

func (m *mockVerifier) IsProposer(id []byte, height, round uint64) bool {
	if m.isProposerFn != nil {
		return m.isProposerFn(id, height, round)
	}

	return false
}

func (m *mockVerifier) IsValidator(from []byte) bool {
	if m.isValidatorFn != nil {
		return m.isValidatorFn(from)
	}

	return false
}
