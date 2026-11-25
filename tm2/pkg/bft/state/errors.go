package state

import "fmt"

type (
	InvalidBlockError error
	ProxyAppConnError error

	AppBlockHeightTooHighError struct {
		CoreHeight int64
		AppHeight  int64
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

func (e AppBlockHeightTooHighError) Error() string {
	return fmt.Sprintf("App block height (%d) is higher than core (%d)", e.AppHeight, e.CoreHeight)
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
