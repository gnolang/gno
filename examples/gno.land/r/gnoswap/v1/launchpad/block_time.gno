package launchpad

type BlockTimeInfo struct {
	blockTime   int64
	blockHeight int64
}

func (bt *BlockTimeInfo) BlockTime() int64 {
	return bt.blockTime
}

func (bt *BlockTimeInfo) setBlockTime(time int64) {
	bt.blockTime = time
}

func (bt *BlockTimeInfo) BlockHeight() int64 {
	return bt.blockHeight
}

func (bt *BlockTimeInfo) setBlockHeight(height int64) {
	bt.blockHeight = height
}

func (bt *BlockTimeInfo) updateHeightByAverageBlockTimeMs(currentTime, currentHeight, averageBlockTimeMs int64) {
	if averageBlockTimeMs <= 0 {
		panic("average block time must be greater than 0")
	}

	if bt.blockTime < currentTime {
		return
	}

	if bt.blockTime == currentTime {
		bt.blockHeight = currentHeight
		
		return
	}

	remainingTimeMs := (bt.blockTime - currentTime) * 1000
	remainingHeight := remainingTimeMs / averageBlockTimeMs
	bt.blockHeight = currentHeight + remainingHeight
}

func newBlockTimeInfo(time, height int64) *BlockTimeInfo {
	return &BlockTimeInfo{
		blockTime:   time,
		blockHeight: height,
	}
}

func newBlockTimeInfoByTimeAndAverageBlockTimeMs(currentTime, currentHeight, targetTime, averageBlockTimeMs int64) *BlockTimeInfo {
	blockInfo := newBlockTimeInfo(targetTime, 0)
	blockInfo.updateHeightByAverageBlockTimeMs(currentTime, currentHeight, averageBlockTimeMs)

	return blockInfo
}
