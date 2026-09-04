package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
	integ "github.com/gnolang/gno/misc/gnoe2e/internal/integration"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newDefaultsCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "defaults",
			ShortUsage: "defaults [config|genesis]",
			ShortHelp:  "list the cluster keys a scenario can declare, with their values",
			LongHelp: "Lists every key a scenario can put in its \"-- cluster --\" section, " +
				"each with the value a cluster boots with when the scenario leaves it alone. " +
				"The values are the harness's, not the chain's: a local cluster commits on " +
				"timings of its own. Naming config or genesis prints that family alone.",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execDefaults(io, args)
		},
	)
}

func execDefaults(io commands.IO, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("expected at most one family, got %d arguments", len(args))
	}

	family := ""
	if len(args) == 1 {
		family = args[0]
	}

	switch family {
	case "":
		printNamedKeys(io)
		io.Println("")
		printGenesisDefaults(io)
		io.Println("")
		printConfigDefaults(io)
	case "genesis":
		printGenesisDefaults(io)
	case "config":
		printConfigDefaults(io)
	default:
		return fmt.Errorf("expected %q, %q or no family, got %q", "config", "genesis", family)
	}
	return nil
}

func printNamedKeys(io commands.IO) {
	io.Println("# Named keys, sugar over the genesis params below. They are applied first, so")
	io.Println("# a scenario that spells out the path a key covers gets the path.")
	io.Println("#")
	io.Println("# validators is required and has no default: a scenario states the cluster it")
	io.Println("# needs. Left unset, code-submission-policy leaves the chain's own policy and")
	io.Println("# pkg-approver leaves the oracle, which is the approver an inert chain gets.")
	io.Println("validators:")
	io.Println("code-submission-policy:")
	io.Println("pkg-approver:")
	io.Printfln("block-max-gas: %d", cluster.DefaultGenesisConfig().MaxGas)
}

func printGenesisDefaults(io commands.IO) {
	io.Println("# Genesis params, keyed as `gnogenesis params set` keys them.")
	for _, o := range cluster.GenesisDefaults() {
		io.Println(clusterLine(integ.GenesisParamsPrefix+o.Key, o.Value))
	}
}

func printConfigDefaults(io commands.IO) {
	io.Println("# Node config, keyed as `gnoland config set` keys it. Applied to every")
	io.Println("# validator after the harness's own passes, so a scenario can have the")
	io.Println("# timings and limits it asks for rather than the ones below.")
	for _, o := range cluster.ConfigDefaults() {
		if slices.Contains(integ.HarnessAssignedConfigKeys, o.Key) {
			io.Printfln("%s%s: (cannot be set: %s)", integ.NodeConfigPrefix, o.Key, integ.HarnessAssignedReason)
			continue
		}
		io.Println(clusterLine(integ.NodeConfigPrefix+o.Key, o.Value))
	}
}

// clusterLine writes one key the way a scenario writes it, which for a setting
// that holds nothing is the bare key.
func clusterLine(key, value string) string {
	if value == "" {
		return key + ":"
	}
	return key + ": " + value
}
