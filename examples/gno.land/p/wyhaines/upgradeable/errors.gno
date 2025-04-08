package upgradeable

import "errors"

// Events emitted by the upgradeable package
const (
	// Function-related events
	EventFunctionRegistered = "FunctionRegistered"
	EventFunctionRemoved    = "FunctionRemoved"
	
	// Proxy-related events
	EventImplementationUpgraded = "ImplementationUpgraded"
	EventProxyInitialized       = "ProxyInitialized"
)

// Errors returned by the upgradeable package
var (
	// Registry errors
	ErrFunctionNotRegistered = errors.New("upgradeable: function not registered")
	ErrInvalidFunction       = errors.New("upgradeable: invalid function provided")
	ErrTypeMismatch          = errors.New("upgradeable: function type mismatch")
	
	// Proxy errors
	ErrImplementationNotSet        = errors.New("upgradeable: implementation not set")
	ErrInvalidImplementation       = errors.New("upgradeable: invalid implementation")
	ErrImplementationCallFailed    = errors.New("upgradeable: implementation call failed")
	ErrProxyAlreadyInitialized     = errors.New("upgradeable: proxy already initialized")
	
	// Security errors
	ErrCallerNotImplementation     = errors.New("upgradeable: caller is not the implementation")
	ErrCallerNotAdmin              = errors.New("upgradeable: caller is not the admin")
)
