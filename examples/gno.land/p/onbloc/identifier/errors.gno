package identifier

import "errors"

var (
	ErrSpoofedRealm = errors.New("identifier: rlm does not match the current crossing frame")
	ErrNotRealm     = errors.New("identifier: rlm must be a realm (got EOA/origin)")
)
