package core

import "testing"

func TestMessageQueue_ConsensusIntegration(t *testing.T) {
	// Create a new instance of MessageQueue
	messageQueue := NewMessageQueue()

	// Define messages from multiple sources with different states
	message1 := &Msg{
		state:  Propose,
		height: 0,
		round:  0,
		block:  []byte{1, 2, 3},
		from:   []byte{0, 1, 2},
	}
	message2 := &Msg{
		state:  Prevote,
		height: 0,
		round:  0,
		block:  []byte{1, 2, 3},
		from:   []byte{0, 1, 2},
	}
	message3 := &Msg{
		state:  Precommit,
		height: 0,
		round:  0,
		block:  []byte{1, 2, 3},
		from:   []byte{0, 1, 2},
	}

	// Add messages to the queue
	messageQueue.AddMessage(message2)
	messageQueue.AddMessage(message3)
	messageQueue.AddMessage(message1)

	// Create a new instance of Tendermint (consensus algorithm)
	tendermint := NewTendermint(&Config{
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
				1: true, // Height 1 is valid
			},
		},
	})

	err := tendermint.Init()

	if err != nil {
		t.Errorf("Error init: %v", err)
	}

	var done bool

	// Process messages from the queue using the consensus algorithm
	for {
		// Pop the next message from the queue
		message := messageQueue.PopMessage()
		if message == nil {
			// No more messages in the queue, exit the loop
			break
		}

		err, done = tendermint.ProcessMsg(message)
		if err != nil {
			t.Errorf("Error processing message: %v", err)
		}
	}

	if !done {
		t.Error("Consensus should have been reached")
	}
}
