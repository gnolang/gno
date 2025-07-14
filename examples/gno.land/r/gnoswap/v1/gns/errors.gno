package gns

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

var (
	errInvalidYear                 = errors.New("[GNOSWAP-GNS-001] invalid year")
	errTooManyEmission             = errors.New("[GNOSWAP-GNS-002] too many emission reward")
	errCallbackEmissionChangeIsNil = errors.New("[GNOSWAP-GNS-003] callback emission change is nil")
	errInvalidAvgBlockTimeInMs     = errors.New("[GNOSWAP-GNS-004] invalid avg block time in ms")
)

func addDetailToError(err error, detail string) string {
	finalErr := ufmt.Errorf("%s || %s", err.Error(), detail)
	return finalErr.Error()
}
