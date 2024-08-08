package core

import (
	"errors"
)

var (
	ErrUnauthorized                = errors.New("unauthorized")
	ErrModuleDisabledCannotExecute = errors.New("module disabled, cannot execute")
	ErrNotImplemented              = errors.New("not implemented")
	ErrModuleAlreadyDisabled       = errors.New("module already disabled")
	ErrNoActiveProposalModules     = errors.New("no active proposal modules")
	ErrModuleAlreadyAdded          = errors.New("module already added")
	ErrNotSupported                = errors.New("not supported")
)
