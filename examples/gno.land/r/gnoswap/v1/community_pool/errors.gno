package community_pool

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

var (
	errNoPermission      = errors.New("[GNOSWAP-COMMUNITY_POOL-001] caller has no permission")
	errNotRegistered     = errors.New("[GNOSWAP-COMMUNITY_POOL-002] not registered")
	errAlreadyRegistered = errors.New("[GNOSWAP-COMMUNITY_POOL-003] already registered")
	errLocked            = errors.New("[GNOSWAP-COMMUNITY_POOL-004] can't transfer token while locked")
	errHalted            = errors.New("[GNOSWAP-COMMUNITY_POOL-005] halted")
)

func addDetailToError(err error, detail string) string {
	finalErr := ufmt.Errorf("%s || %s", err.Error(), detail)
	return finalErr.Error()
}
