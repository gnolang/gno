package gnoland

import (
	"regexp"

	"github.com/gnolang/gno/gnovm/stdlibs/chain"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/events"
)

const (
	valRealm     = "gno.land/r/sys/validators/v2" // XXX: make it configurable from GovDAO
	valChangesFn = "GetChanges"

	validatorAddedEvent   = "ValidatorAdded"
	validatorRemovedEvent = "ValidatorRemoved"
)

// XXX: replace with amino-based clean approach
var valRegexp = regexp.MustCompile(`{\("([^"]*)"\s[^)]+\),\("((?:[^"]|\\")*)"\s[^)]+\),\((\d+)\s[^)]+\)}`)

// validatorUpdate is a type being used for "notifying"
// that a validator change happened on-chain. The events from `r/sys/validators`
// do not pass data related to validator add / remove instances (who, what, how)
type validatorUpdate struct{}

// validatorEventFilter filters the given event to determine if it
// is tied to a validator update
func validatorEventFilter(event events.Event) []validatorUpdate {
	// Make sure the event is a new TX event
	txResult, ok := event.(types.EventTx)
	if !ok {
		return nil
	}

	// Make sure an add / remove event happened
	for _, ev := range txResult.Result.Response.Events {
		// Make sure the event is a GnoVM event
		gnoEv, ok := ev.(chain.Event)
		if !ok {
			continue
		}

		// Make sure the event is from `r/sys/validators`
		if gnoEv.PkgPath != valRealm {
			continue
		}

		// Make sure the event is either an add / remove
		switch gnoEv.Type {
		case validatorAddedEvent, validatorRemovedEvent:
			// We don't pass data around with the events, but a single
			// notification is enough to "trigger" a VM scrape
			return []validatorUpdate{{}}
		default:
			continue
		}
	}

	return nil
}
