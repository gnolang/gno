package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/file"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/null"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type configEventsCfg struct {
	commonEditCfg

	eventStoreType   string
	eventStoreParams commands.StringArr
}

// newConfigEventsCmd creates the new config events command
func newConfigEventsCmd(io commands.IO) *commands.Command {
	cfg := &configEventsCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "events",
			ShortUsage: "config events [flags]",
			ShortHelp:  "Edits the Gno node's event store configuration",
			LongHelp:   "Edits the Gno node's event store configuration locally",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execConfigEvents(cfg, io)
		},
	)

	return cmd
}

func (c *configEventsCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonEditCfg.RegisterFlags(fs)

	fs.StringVar(
		&c.eventStoreType,
		"event-store-type",
		null.EventStoreType,
		fmt.Sprintf(
			"type of transaction event store [%s]",
			strings.Join(
				[]string{
					null.EventStoreType,
					file.EventStoreType,
				},
				", ",
			),
		),
	)

	fs.Var(
		&c.eventStoreParams,
		"event-store-params",
		"the params for the event store, in the form <key>=<value>",
	)
}

func execConfigEvents(cfg *configEventsCfg, io commands.IO) error {
	// Load the config
	loadedCfg, err := config.LoadConfigFile(cfg.configPath)
	if err != nil {
		return fmt.Errorf("unable to load config, %w", err)
	}

	// Set the event store type, if any
	if cfg.eventStoreType != "" {
		loadedCfg.TxEventStore.EventStoreType = cfg.eventStoreType
	}

	// Set the event store params, if any
	if len(cfg.eventStoreParams) > 0 {
		loadedCfg.TxEventStore.Params = parseEventStoreParams(cfg.eventStoreParams)
	}

	// Save the config
	if err := config.WriteConfigFile(cfg.configPath, loadedCfg); err != nil {
		return fmt.Errorf("unable to save updated config, %w", err)
	}

	io.Printfln("Updated event store configuration saved at %s", cfg.configPath)

	return nil
}

// parseEventStoreParams parses the event store params into
// a param map
func parseEventStoreParams(values []string) types.EventStoreParams {
	params := make(types.EventStoreParams, len(values))

	for _, pair := range values {
		// Split the string into key and value
		kv := strings.SplitN(pair, "=", 2)

		// Check if the split produced exactly two elements
		if len(kv) != 2 {
			continue
		}

		key := kv[0]
		value := kv[1]

		params[key] = value
	}

	return params
}
