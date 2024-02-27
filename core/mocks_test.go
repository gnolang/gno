package core

import "github.com/gnolang/go-tendermint/messages/types"

type broadcastDelegate func(*types.Message)

type mockBroadcast struct {
	broadcastFn broadcastDelegate
}

func (m *mockBroadcast) Broadcast(message *types.Message) {
	if m.broadcastFn != nil {
		m.broadcastFn(message)
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

type signDelegate func([]byte) []byte

type mockSigner struct {
	signFn signDelegate
}

func (m *mockSigner) Sign(data []byte) []byte {
	if m.signFn != nil {
		return m.signFn(data)
	}

	return nil
}
