package state

import "fmt"

type (
	InvalidBlockError error
	ProxyAppConnError error

	UnknownBlockError struct {
		Height int64
	}

	BlockHashMismatchError struct {
		CoreHash []byte
		AppHash  []byte
		Height   int64
	}

	AppBlockHeightTooHighError struct {
		CoreHeight int64
		AppHeight  int64
	}

	LastStateMismatchError struct {
		Height int64
		Core   []byte
		App    []byte
	}

	StateMismatchError struct {
		Got      *State
		Expected *State
	}

	NoValSetForHeightError struct {
		Height int64
	}

	NoConsensusParamsForHeightError struct {
		Height int64
	}

	NoABCIResponsesForHeightError struct {
		Height int64
	}

	NoTxResultForHashError struct {
		Hash []byte
	}
)

func (e UnknownBlockError) Error() string {
	return fmt.Sprintf("Could not find block #%d", e.Height)
}

func (e BlockHashMismatchError) Error() string {
	return fmt.Sprintf("App block hash (%X) does not match core block hash (%X) for height %d", e.AppHash, e.CoreHash, e.Height)
}

func (e AppBlockHeightTooHighError) Error() string {
	return fmt.Sprintf("App block height (%d) is higher than core (%d)", e.AppHeight, e.CoreHeight)
}

func (e LastStateMismatchError) Error() string {
	return fmt.Sprintf("Latest tendermint block (%d) LastAppHash (%X) does not match app's AppHash (%X)", e.Height, e.Core, e.App)
}

func (e StateMismatchError) Error() string {
	return fmt.Sprintf("State after replay does not match saved state. Got ----\n%v\nExpected ----\n%v\n", e.Got, e.Expected)
}

func (e NoValSetForHeightError) Error() string {
	return fmt.Sprintf("Could not find validator set for height #%d", e.Height)
}

func (e NoConsensusParamsForHeightError) Error() string {
	return fmt.Sprintf("Could not find consensus params for height #%d", e.Height)
}

func (e NoABCIResponsesForHeightError) Error() string {
	return fmt.Sprintf("Could not find results for height #%d", e.Height)
}

func (e NoTxResultForHashError) Error() string {
	return fmt.Sprintf("Could not find tx result for hash #%X", e.Hash)
}
