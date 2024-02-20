package core

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type blMock struct {
	validBlocks map[uint64]bool
	isMajority  bool
}

func (b *blMock) IsValid(height uint64, block []byte) bool {
	valid, ok := b.validBlocks[height]
	if !ok {
		// Default to true if height is not specified in the map
		return true
	}
	return valid
}

func (b *blMock) IsMajority(votes int) bool {
	return b.isMajority
}

type brMock struct{}

func (b *brMock) Broadcast(state State, height uint64, round int64, block []byte) {}

type prMock struct {
	getFunc func(height uint64, block []byte, round int64) []byte
	me      []byte
}

func (b *prMock) Get(height uint64, block []byte, round int64) []byte {
	return b.getFunc(height, block, round)
}

func (b *prMock) Me() []byte {
	return b.me
}

func TestTendermint_ProcessMsgOk(t *testing.T) {
	tm := NewTendermint(&Config{
		ctx:    nil,
		height: 0,
		p: &prMock{getFunc: func(height uint64, block []byte, round int64) []byte {
			return []byte{0, 1, 2}
		}, me: []byte{0, 1, 2}},
		b:       &brMock{},
		timeout: 0,
		block:   nil,
		bv: &blMock{
			isMajority: true,
			validBlocks: map[uint64]bool{
				1: true,  // Height 1 is valid
				2: false, // Height 2 is invalid
			},
		},
	})

	err := tm.Init()

	if err != nil {
		panic(err)
	}

	err, finished := tm.ProcessMsg(&Msg{
		state:  Propose,
		height: 0,
		round:  0,
		block:  []byte{1, 2, 3},
		from:   []byte{0, 1, 2},
	})

	if err != nil {
		panic(err)
	}

	err, finished = tm.ProcessMsg(&Msg{
		state:  Prevote,
		height: 0,
		round:  0,
		block:  []byte{1, 2, 3},
		from:   []byte{0, 1, 2},
	})

	if err != nil {
		panic(err)
	}

	if finished {
		panic("should not have finish as true on prevote")
	}

	if tm.state != Precommit {
		panic(fmt.Sprintf("expected state to be Precommit after majority Prevote, got: %+v", tm.state))
	}

	err, finished = tm.ProcessMsg(&Msg{
		state:  Precommit,
		height: 0,
		round:  0,
		block:  []byte{1, 2, 3},
		from:   []byte{0, 1, 2},
	})

	if err != nil {
		panic(err)
	}

	if !finished {
		panic("should have finish as true on majority Precommit")
	}

	if tm.state != Precommit {
		panic(fmt.Sprintf("expected state to be Precommit after majority Precommit, got: %+v", tm.state))
	}
}

func TestTendermint_ProcessMsg(t *testing.T) {
	cfg := &Config{
		ctx:    nil,
		height: 1,
		p: &prMock{
			getFunc: func(height uint64, block []byte, round int64) []byte {
				// Mock behavior for proposer get function
				return []byte("MockProposer")
			},
			me: []byte("MockProposer"),
		},
		b:       &brMock{},
		timeout: time.Second,
		block:   []byte("block"),
		bv: &blMock{
			validBlocks: map[uint64]bool{
				1: true,  // Height 1 is valid
				2: false, // Height 2 is invalid
			},
			isMajority: true,
		},
	}

	// Create a new Tendermint instance
	tendermint := NewTendermint(cfg)

	// Initialize Tendermint
	err := tendermint.Init()
	if err != nil {
		t.Errorf("Error initializing Tendermint: %v", err)
	}

	// Define test scenarios
	tests := []struct {
		name         string
		mockMsg      *Msg
		expectedErr  error
		expectedDone bool
	}{
		{
			name: "Valid Proposal",
			mockMsg: &Msg{
				state:  Propose,
				height: 1,
				round:  0,
				block:  []byte("block"),
				from:   []byte("MockProposer"),
			},
			expectedErr:  nil,
			expectedDone: false,
		},
		{
			name: "Invalid Block",
			mockMsg: &Msg{
				state:  Propose,
				height: 2,
				round:  0,
				block:  []byte("invalid_block"),
				from:   []byte("MockProposer"),
			},
			expectedErr:  errors.New("invalid height: 2 block: [105 110 118 97 108 105 100 95 98 108 111 99 107]\n"),
			expectedDone: false,
		},
		// Add more test scenarios as needed
	}

	// Run test scenarios
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err, done := tendermint.ProcessMsg(test.mockMsg)
			if (err != nil && test.expectedErr == nil) || (err == nil && test.expectedErr != nil) || (err != nil && test.expectedErr != nil && err.Error() != test.expectedErr.Error()) {
				t.Errorf("Error mismatch for test '%s'. Expected: %v, Got: %v", test.name, test.expectedErr, err)
			}
			if done != test.expectedDone {
				t.Errorf("Done status mismatch for test '%s'. Expected: %v, Got: %v", test.name, test.expectedDone, done)
			}
		})
	}
}
