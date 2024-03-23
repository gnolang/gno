package checkpoints

import (
	std "std"
)

type Checkpoint struct {
	blockNumber int64
	value       uint64
}

func NewCheckPoint(blockNumber int64, value uint64) Checkpoint {
	return Checkpoint{
		blockNumber: blockNumber,
		value:       value,
	}
}

func (ck Checkpoint) GetBlockNumber() int64 {
	return ck.blockNumber
}

func (ck Checkpoint) GetValue() uint64 {
	return ck.value
}

type History struct {
	Checkpoints []Checkpoint
}

func NewHistory() *History {
	return &History{}
}

// just ordered entries, no summation
func (h *History) UpdateOrderedHistory(ck Checkpoint) {
	if h.LatestTime() > ck.blockNumber {
		panic("not ordered, should not happen")
	}
	h.Checkpoints = append(h.Checkpoints, ck)
}

func (history *History) LatestTime() int64 {
	pos := len(history.Checkpoints)
	if pos == 0 {
		return 0
	} else {
		return history.Checkpoints[pos-1].blockNumber
	}
}

// Returns the value in the latest checkpoint, or zero if there are no checkpoints.
func (history *History) LatestValue() uint64 {
	pos := len(history.Checkpoints)
	if pos == 0 {
		return 0
	} else {
		return history.Checkpoints[pos-1].value
	}
}

func (history *History) IterateHistory(height int64, cb func(ck Checkpoint)) {
	for _, ck := range history.Checkpoints {
		// mature
		if ck.blockNumber <= height {
			cb(ck)
		} else {
			break
		}
	}
}

// Returns the value at a given block number.
// If a checkpoint is not available at that block, the closest one before it is returned, or zero otherwise.

// xxx: snapshot happens when token transferred. if(say), voter doesn't send or receive token at height 100, then there would
// not be any record in the checkpoint at height 100, but it doesn't mean the voter is not qualified, he may have a snapshot
// at 99, 98 or any height <100, which means the nearest snapshot is also counted.
func (history *History) GetAtBlock(blockNumber int64) uint64 {
	height := std.GetHeight()
	if blockNumber >= height { // only mined blocks
		panic("Checkpoints: block not yet mined")
	}
	var high int64
	var low int64

	high = int64(len(history.Checkpoints))
	low = 0
	for low < high {
		mid := _average(low, high)
		if history.Checkpoints[mid].blockNumber > blockNumber {
			high = mid
		} else {
			low = mid + 1
		}
	}
	if high == 0 {
		return 0
	}
	return history.Checkpoints[high-1].value
}

// (a + b) / 2 can overflow.
func _average(a, b int64) int64 {
	return (a & b) + (a^b)/2
}

// Pushes a value onto a History so that it is stored as the checkpoint for the current block.
// Returns previous value and new value.
func (history *History) Push(value uint64) (uint64, uint64) {
	pos := len(history.Checkpoints)
	old := history.LatestValue()
	height := std.GetHeight()
	// same block, update
	if pos > 0 && history.Checkpoints[pos-1].blockNumber == height {
		history.Checkpoints[pos-1].value = value
	} else {
		history.Checkpoints = append(history.Checkpoints, Checkpoint{blockNumber: height, value: value})
	}
	return old, value
}

// Pushes a value onto a History, by updating the latest value using binary operation `op`. The new value will
// be set to `op(latest, delta)`.
// Returns previous value and new value.
func (history *History) PushWithOp(op func(uint64, uint64) uint64, delta uint64) (uint64, uint64) {
	return history.Push(op(history.LatestValue(), delta))
}
