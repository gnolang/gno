package core

import "sort"

// MessageQueue represents a message queue that maintains messages sorted by state
// This is helpful when some nodes are ahead (in terms of state) of others
type MessageQueue struct {
	messages []*Msg
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		messages: make([]*Msg, 0),
	}
}

// AddMessage adds a message to the queue while maintaining the sorted order by state
func (mq *MessageQueue) AddMessage(msg *Msg) {
	index := sort.Search(len(mq.messages), func(i int) bool {
		return mq.messages[i].state >= msg.state
	})

	mq.messages = append(mq.messages[:index], append([]*Msg{msg}, mq.messages[index:]...)...)
}

// PopMessage pops and returns the next message that should be processed
func (mq *MessageQueue) PopMessage() *Msg {
	if len(mq.messages) == 0 {
		return nil
	}
	// The next message to be processed is the one with the lowest state
	return mq.popMessageByState(mq.messages[0].state)
}

func (mq *MessageQueue) popMessageByState(state State) *Msg {
	for i, msg := range mq.messages {
		if msg.state == state {
			result := msg
			mq.messages = append(mq.messages[:i], mq.messages[i+1:]...)
			return result
		}
	}
	return nil
}
