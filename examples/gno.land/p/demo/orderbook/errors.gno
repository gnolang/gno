package orderbook

import "errors"

var (
	ErrFailedToAddOrder    = errors.New("failed to add order")
	ErrFailedToRemoveOrder = errors.New("failed to remove order")
	ErrFailedToPeekOrder   = errors.New("failed to peek order")
	ErrUnknownOrderType    = errors.New("unknown order type")
)
