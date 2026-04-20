package gnoland

import (
	"regexp"

	"github.com/gnolang/gno/gnovm/stdlibs/chain"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

const (
	valRealm     = "gno.land/r/sys/validators/v2" // XXX: make it configurable from GovDAO
	valChangesFn = "GetChanges"

	validatorAddedEvent   = "ValidatorAdded"
	validatorRemovedEvent = "ValidatorRemoved"
)

// XXX: replace with amino-based clean approach
var valRegexp = regexp.MustCompile(`{\("([^"]*)"\s[^)]+\),\("((?:[^"]|\\")*)"\s[^)]+\),\((\d+)\s[^)]+\)}`)

// hasValidatorChangeEvent reports whether any event in evs is a validator
// add or remove event from the validators realm.
func hasValidatorChangeEvent(evs []abci.Event) bool {
	for _, ev := range evs {
		gnoEv, ok := ev.(chain.Event)
		if !ok {
			continue
		}
		if gnoEv.PkgPath != valRealm {
			continue
		}
		switch gnoEv.Type {
		case validatorAddedEvent, validatorRemovedEvent:
			return true
		}
	}
	return false
}
