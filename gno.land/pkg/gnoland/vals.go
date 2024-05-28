package gnoland

import (
	"fmt"
	"strconv"

	gnovm "github.com/gnolang/gno/gnovm/stdlibs/std"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/events"
)

const (
	valRealm = "gno.land/r/sys/vals"

	validatorAddedEvent   = "ValidatorAdded"
	validatorRemovedEvent = "ValidatorRemoved"

	addressEventKey     = "address"
	pubKeyEventKey      = "pub_key"
	votingPowerEventKey = "voting_power"
)

// validatorEventFilter
func validatorEventFilter(event events.Event) []abci.ValidatorUpdate {
	// Make sure the event is a new TX event
	txResult, ok := event.(types.EventTx)
	if !ok {
		return nil
	}

	// extractValUpdate parses the event attributes and extracts the relevant
	// validator change data
	extractValUpdate := func(attributes []gnovm.GnoEventAttribute) (*abci.ValidatorUpdate, error) {
		// Extract the event attributes
		attrs := extractEventAttributes(attributes)

		var (
			addressRaw     = attrs[addressEventKey]
			pubKeyRaw      = attrs[pubKeyEventKey]
			votingPowerRaw = attrs[votingPowerEventKey]
		)

		// Parse the address
		address, err := crypto.AddressFromBech32(addressRaw)
		if err != nil {
			return nil, fmt.Errorf("unable to parse address, %w", err)
		}

		// Parse the public key
		pubKey, err := crypto.PubKeyFromBech32(pubKeyRaw)
		if err != nil {
			return nil, fmt.Errorf("unable to parse public key, %w", err)
		}

		// Parse the voting power
		votingPower, err := strconv.Atoi(votingPowerRaw)
		if err != nil {
			return nil, fmt.Errorf("unable to parse voting power, %w", err)
		}

		return &abci.ValidatorUpdate{
			Address: address,
			PubKey:  pubKey,
			Power:   int64(votingPower),
		}, nil
	}

	// Extract the validator change events
	valUpdates := make([]abci.ValidatorUpdate, 0)
	for _, ev := range txResult.Result.Response.Events {
		// Make sure the event is a GnoVM event
		gnoEv, ok := ev.(gnovm.GnoEvent)
		if !ok {
			continue
		}

		// Make sure the event is from `r/sys/vals`
		if gnoEv.PkgPath != valRealm {
			continue
		}

		// Make sure the event is either an add / remove
		switch gnoEv.Type {
		case validatorAddedEvent:
			update, err := extractValUpdate(gnoEv.Attributes)
			if err != nil {
				continue
			}

			valUpdates = append(valUpdates, *update)
		case validatorRemovedEvent:
			update, err := extractValUpdate(gnoEv.Attributes)
			if err != nil {
				continue
			}

			// Validator updates that have Power == 0
			// are considered to be "remove" signals
			update.Power = 0

			valUpdates = append(valUpdates, *update)
		default:
			continue
		}
	}

	return valUpdates
}

// extractEventAttributes generates an attribute map from
// the gno event attributes, for quick lookup
func extractEventAttributes(evAttrs []gnovm.GnoEventAttribute) map[string]string {
	attrs := make(map[string]string, len(evAttrs))

	for _, attr := range evAttrs {
		attrs[attr.Key] = attr.Value
	}

	return attrs
}
