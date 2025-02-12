package protocol_fee

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

var (
	errNoPermission  = errors.New("[GNOSWAP-PROTOCOL_FEE-001] caller has no permission")
	errInvalidPct    = errors.New("[GNOSWAP-PROTOCOL_FEE-002] invalid percentage")
	errInvalidAmount = errors.New("[GNOSWAP-PROTOCOL_FEE-003] invalid amount")
)

func addDetailToError(err error, detail string) string {
	finalErr := ufmt.Errorf("%s || %s", err.Error(), detail)
	return finalErr.Error()
}
