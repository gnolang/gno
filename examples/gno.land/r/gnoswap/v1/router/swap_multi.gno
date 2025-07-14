package router

import (
	"std"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

// SwapDirection represents the direction of swap execution in multi-hop swaps.
// It determines whether swaps are processed in forward order (first to last pool)
// or backward order (last to first pool).
type SwapDirection int

const (
	_ SwapDirection = iota
	// Forward indicates a swap processing direction from the first pool to the last pool.
	// Used primarily for exactIn swaps where the input amount is known.
	Forward

	// Backward indicates a swap processing direction from the last pool to the first pool.
	// Used primarily for exactOut swaps where the output amount is known and input amounts
	// need to be calculated in reverse order.
	Backward
)

// SwapExecutor defines the interface for swap operation execution
type MultiSwapExecutor interface {
	// Run performs the swap operation and returns pool received and pool output amounts
	Run(p SwapParams, data SwapCallbackData, recipient std.Address) (*u256.Uint, *u256.Uint)
}

// DrySwapExecutor implements SwapExecutor for dry run simulations
type DryMultiSwapExecutor struct{}

// Execute performs a dry swap operation without changing state
func (e *DryMultiSwapExecutor) Run(p SwapParams, data SwapCallbackData, _ std.Address) (*u256.Uint, *u256.Uint) {
	return swapDryInner(cross, p.amountSpecified, zero, data)
}

// RealSwapExecutor implements SwapExecutor for actual swap operations
type RealMultiSwapExecutor struct{}

// Execute performs a real swap operation with state changes
func (e *RealMultiSwapExecutor) Run(p SwapParams, data SwapCallbackData, recipient std.Address) (*u256.Uint, *u256.Uint) {
	return swapInner(cross, p.amountSpecified, recipient, zero, data)
}

// SwapProcessor handles the execution flow for multi-hop swaps
type MultiSwapProcessor struct {
	executor   MultiSwapExecutor
	direction  SwapDirection
	isSimulate bool
}

var (
	_ MultiSwapExecutor = (*DryMultiSwapExecutor)(nil)
	_ MultiSwapExecutor = (*RealMultiSwapExecutor)(nil)
)

// NewSwapProcessor creates a new SwapProcessor with the specified configuration
func NewMultiSwapProcessor(isSimulate bool, direction SwapDirection) *MultiSwapProcessor {
	var executor MultiSwapExecutor
	if isSimulate {
		executor = &DryMultiSwapExecutor{}
	} else {
		executor = &RealMultiSwapExecutor{}
	}

	return &MultiSwapProcessor{
		executor:   executor,
		direction:  direction,
		isSimulate: isSimulate,
	}
}

// processForwardSwap handles forward direction swaps (exactIn)
func (p *MultiSwapProcessor) processForwardSwap(sp SwapParams, numPools int, swapPath string) (*u256.Uint, *u256.Uint, error) {
	firstAmountIn := zero
	payer := getPrevAddr() // Initial payer is the user
	currentPoolIndex := 0

	for {
		var recipient std.Address
		currentPoolIndex++

		// Determine the recipient for this hop
		if currentPoolIndex < numPools {
			recipient = routerAddr
		} else {
			recipient = sp.recipient
		}

		// Execute the swap operation
		callbackData := newSwapCallbackData(sp, payer)
		amountIn, amountOut := p.executor.Run(sp, callbackData, recipient)

		// Record the first hop's input amount
		if currentPoolIndex == 1 {
			firstAmountIn = amountIn
		}

		// Check if we've processed all hops
		if currentPoolIndex >= numPools {
			if p.isSimulate {
				return firstAmountIn, amountOut, nil
			}
			return firstAmountIn, amountOut, nil
		}

		// Update parameters for the next hop
		payer = routerAddr
		nextInput, nextOutput, nextFee := getDataForMultiPath(swapPath, currentPoolIndex)
		sp.tokenIn = nextInput
		sp.tokenOut = nextOutput
		sp.fee = nextFee
		sp.amountSpecified = i256.FromUint256(amountOut)
	}
}

// processBackwardSwap handles backward direction swaps (exactOut)
func (p *MultiSwapProcessor) processBackwardSwap(sp SwapParams, numPools int, swapPath string) (*u256.Uint, *u256.Uint, error) {
	if !p.isSimulate {
		return p.processBackwardRealSwap(sp, numPools, swapPath)
	}
	return p.processBackwardDrySwap(sp, numPools, swapPath)
}

// processBackwardDrySwap handles backward simulated swaps
func (p *MultiSwapProcessor) processBackwardDrySwap(sp SwapParams, numPools int, swapPath string) (*u256.Uint, *u256.Uint, error) {
	firstAmountIn := u256.Zero()
	currentPoolIndex := numPools - 1
	payer := routerAddr

	for {
		callbackData := newSwapCallbackData(sp, payer)
		amountIn, amountOut := p.executor.Run(sp, callbackData, sp.recipient)

		if currentPoolIndex == 0 {
			firstAmountIn = amountIn
		}

		currentPoolIndex--

		if currentPoolIndex == -1 {
			return firstAmountIn, amountOut, nil
		}

		// Update parameters for the next hop
		nextInput, nextOutput, nextFee := getDataForMultiPath(swapPath, currentPoolIndex)
		intAmountIn := i256.FromUint256(amountIn)

		sp.amountSpecified = i256.Zero().Neg(intAmountIn)
		sp.tokenIn = nextInput
		sp.tokenOut = nextOutput
		sp.fee = nextFee
	}
}

// processBackwardRealSwap handles backward real swaps
func (p *MultiSwapProcessor) processBackwardRealSwap(sp SwapParams, numPools int, swapPath string) (*u256.Uint, *u256.Uint, error) {
	// First collect all swap information by simulating backward
	swapInfo := p.collectBackwardSwapInfo(sp, numPools, swapPath)

	// Then execute swaps in forward order
	return p.executeCollectedSwaps(swapInfo, sp.recipient)
}

// collectBackwardSwapInfo simulates swaps backward to collect parameters
func (p *MultiSwapProcessor) collectBackwardSwapInfo(sp SwapParams, numPools int, swapPath string) []SingleSwapParams {
	swapInfo := []SingleSwapParams{}
	currentPoolIndex := numPools

	for currentPoolIndex >= 0 {
		thisSwap := SingleSwapParams{
			tokenIn:         sp.tokenIn,
			tokenOut:        sp.tokenOut,
			fee:             sp.fee,
			amountSpecified: sp.amountSpecified,
		}

		// dry simulation to calculate input amount
		amountIn, _ := singleDrySwap(&thisSwap)
		swapInfo = append(swapInfo, thisSwap)

		if currentPoolIndex == 0 {
			break
		}
		currentPoolIndex--

		// Update parameters for the next simulation
		nextInput, nextOutput, nextFee := getDataForMultiPath(swapPath, currentPoolIndex)

		sp.tokenIn = nextInput
		sp.tokenOut = nextOutput
		sp.fee = nextFee
		sp.amountSpecified = i256.Zero().Neg(i256.FromUint256(amountIn))
	}

	return swapInfo
}

// executeCollectedSwaps performs the collected swaps in forward order
func (p *MultiSwapProcessor) executeCollectedSwaps(swapInfo []SingleSwapParams, finalRecipient std.Address) (*u256.Uint, *u256.Uint, error) {
	firstAmountIn := zero
	currentPoolIndex := len(swapInfo) - 1
	payer := getPrevAddr() // Initial payer is the user

	for currentPoolIndex >= 0 {
		var recipient std.Address
		if currentPoolIndex == 0 {
			recipient = finalRecipient
		} else {
			recipient = routerAddr
		}

		// Execute the swap
		callbackData := newSwapCallbackData(swapInfo[currentPoolIndex], payer)

		amountIn, amountOut := swapInner(
			cross,
			swapInfo[currentPoolIndex].amountSpecified,
			recipient,
			zero,
			callbackData,
		)

		// Record the first hop's input amount
		if currentPoolIndex == len(swapInfo)-1 {
			firstAmountIn = amountIn
		}

		if currentPoolIndex == 0 {
			return firstAmountIn, amountOut, nil
		}

		// Update parameters for the next swap
		swapInfo[currentPoolIndex-1].amountSpecified = i256.FromUint256(amountOut)
		payer = routerAddr
		currentPoolIndex--
	}

	return firstAmountIn, zero, nil
}

// multiSwap performs a multi-hop swap in forward direction
func multiSwap(p SwapParams, numPools int, swapPath string) (*u256.Uint, *u256.Uint) {
	result, output, _ := NewMultiSwapProcessor(false, Forward).
		processForwardSwap(p, numPools, swapPath)
	return result, output
}

// multiSwapNegative performs a multi-hop swap in backward direction
func multiSwapNegative(p SwapParams, numPools int, swapPath string) (*u256.Uint, *u256.Uint) {
	result, output, _ := NewMultiSwapProcessor(false, Backward).
		processBackwardSwap(p, numPools, swapPath)
	return result, output
}

// multiDrySwap simulates a multi-hop swap in forward direction
func multiDrySwap(p SwapParams, numPool int, swapPath string) (*u256.Uint, *u256.Uint, error) {
	return NewMultiSwapProcessor(true, Forward).
		processForwardSwap(p, numPool, swapPath)
}

// multiDrySwapNegative simulates a multi-hop swap in backward direction
func multiDrySwapNegative(p SwapParams, currentPoolIndex int, swapPath string) (*u256.Uint, *u256.Uint, error) {
	return NewMultiSwapProcessor(true, Backward).
		processBackwardSwap(p, currentPoolIndex+1, swapPath)
}
