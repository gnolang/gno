package errors

import (
	"errors"
)

var (
	// Pool errors
	ErrPoolNotFound = errors.New("GNOLEND-001: pool not found")
	ErrPoolAlreadyExists = errors.New("GNOLEND-002: pool already exists")
	ErrInsufficientLiquidity = errors.New("GNOLEND-003: insufficient liquidity in pool")
	
	// Position errors
	ErrInsufficientCollateral = errors.New("GNOLEND-101: insufficient collateral")
	ErrPositionUnhealthy = errors.New("GNOLEND-102: position would become unhealthy")
	ErrPositionNotLiquidatable = errors.New("GNOLEND-103: position is not liquidatable")
	ErrExceedsMaxLTV = errors.New("GNOLEND-104: exceeds maximum LTV ratio")
	
	// Asset errors
	ErrInvalidAsset = errors.New("GNOLEND-201: invalid asset")
	ErrAssetNotEnabled = errors.New("GNOLEND-202: asset not enabled as collateral")
	ErrInsufficientBalance = errors.New("GNOLEND-203: insufficient balance")
	ErrInvalidAmount = errors.New("GNOLEND-204: invalid amount")
	
	// Price errors
	ErrInvalidPrice = errors.New("GNOLEND-301: invalid price")
	ErrStalePrice = errors.New("GNOLEND-302: price data is stale")
	ErrPriceNotAvailable = errors.New("GNOLEND-303: price not available")
	
	// Permission errors
	ErrUnauthorized = errors.New("GNOLEND-401: unauthorized")
	ErrNotOwner = errors.New("GNOLEND-402: not owner")
	
	// Parameter errors
	ErrInvalidParameter = errors.New("GNOLEND-501: invalid parameter")
	ErrParameterOutOfRange = errors.New("GNOLEND-502: parameter out of range")
	
	// System errors
	ErrSystemPaused = errors.New("GNOLEND-901: system is paused")
	ErrInvalidState = errors.New("GNOLEND-902: invalid system state")
) 