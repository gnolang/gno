package common

import "errors"

var (
	ErrAppStateNotSet      = errors.New("genesis app state not set")
	ErrNoOutputFile        = errors.New("no output file path specified")
	ErrUnableToLoadGenesis = errors.New("unable to load genesis")
)
