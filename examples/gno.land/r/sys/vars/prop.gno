package vars

import (
	"std"

	"gno.land/p/demo/dao"
	govdao "gno.land/r/gov/dao/v2"
)

const (
	KeyAddedEvent   = "KeyAdded"   // emitted when a key was added to the set
	KeyRemovedEvent = "KeyRemoved" // emitted when a key was removed from the set
	KeyUpdatedEvent = "KeyUpdated" // emitted when a key value was updated in the set

	KeyEventNameKey = "name"
)

type KV struct {
	Key   string
	Value interface{}
}

// NewVarPropExecutor creates a new proposal executor for
// changing r/sys/vars values through governance proposals
func NewVarPropExecutor(changesFn func() []KV) dao.Executor {
	cb := func() error {
		for _, kv := range changesFn() {
			if kv.Value == nil {
				// Removal request
				vars.Remove(kv.Key)

				// Emit the key set change
				std.Emit(
					KeyRemovedEvent,
					KeyEventNameKey, kv.Key,
				)

				continue
			}

			// Check if it's an update request
			if vars.Has(kv.Key) {
				// Update the value
				vars.Set(kv.Key, kv.Value)

				// Emit the key set change
				std.Emit(
					KeyUpdatedEvent,
					KeyEventNameKey, kv.Key,
				)

				continue
			}

			// Set the new value
			vars.Set(kv.Key, kv.Value)

			// Emit the key set change
			std.Emit(
				KeyAddedEvent,
				KeyEventNameKey, kv.Key,
			)
		}

		return nil
	}

	return govdao.NewGovDAOExecutor(cb)
}
