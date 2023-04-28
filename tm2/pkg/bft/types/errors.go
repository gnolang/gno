package types

import "fmt"

type (
	// InvalidCommitHeightError is returned when we encounter a commit with an
	// unexpected height.
	InvalidCommitHeightError struct {
		Expected int64
		Actual   int64
	}

	// InvalidCommitPrecommitsError is returned when we encounter a commit where
	// the number of precommits doesn't match the number of validators.
	InvalidCommitPrecommitsError struct {
		Expected int
		Actual   int
	}
)

func NewErrInvalidCommitHeight(expected, actual int64) InvalidCommitHeightError {
	return InvalidCommitHeightError{
		Expected: expected,
		Actual:   actual,
	}
}

func (e InvalidCommitHeightError) Error() string {
	return fmt.Sprintf("Invalid commit -- wrong height: %v vs %v", e.Expected, e.Actual)
}

func NewErrInvalidCommitPrecommits(expected, actual int) InvalidCommitPrecommitsError {
	return InvalidCommitPrecommitsError{
		Expected: expected,
		Actual:   actual,
	}
}

func (e InvalidCommitPrecommitsError) Error() string {
	return fmt.Sprintf("Invalid commit -- wrong set size: %v vs %v", e.Expected, e.Actual)
}
