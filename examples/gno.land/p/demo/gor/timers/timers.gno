package timers

import "std"

type Timestamp struct {
	deadline int64
}

func NewTimestamp() *Timestamp {
	return &Timestamp{
		deadline: 0,
	}
}

func (timer *Timestamp) getDeadline() int64 {
	return timer.deadline
}

func (timer *Timestamp) setDeadline(timestamp int64) {
	timer.deadline = timestamp
}

func (timer *Timestamp) reset() {
	timer.deadline = 0
}

func (timer *Timestamp) isUnset() bool {
	return timer.deadline == 0
}

func (timer *Timestamp) isStarted() bool {
	return timer.deadline > 0
}

// TODO: block.Timestamp, do we need this?
func (timer *Timestamp) isPending() bool {
	return timer.deadline > int64(std.GetTimestamp())
}

func (timer *Timestamp) isExpired() bool {
	return timer.isStarted() && timer.deadline <= int64(std.GetTimestamp())
}

type BlockNumber struct {
	deadline int64
}

func NewBlockNumber() *BlockNumber {
	return &BlockNumber{
		deadline: 0,
	}
}

func (timer *BlockNumber) GetDeadline() int64 {
	return timer.deadline
}

func (timer *BlockNumber) SetDeadline(timestamp int64) {
	timer.deadline = timestamp
}

func (timer *BlockNumber) reset() {
	timer.deadline = 0
}

func (timer *BlockNumber) IsSet() bool {
	return timer.deadline == 0
}

func (timer *BlockNumber) isStarted() bool {
	return timer.deadline > 0
}

func (timer *BlockNumber) isPending() bool {
	height := std.GetHeight()
	return timer.deadline > height
}

func (timer *BlockNumber) isExpired() bool {
	height := std.GetHeight()
	return timer.isStarted() && timer.deadline <= height
}
