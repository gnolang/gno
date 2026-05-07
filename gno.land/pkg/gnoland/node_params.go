package gnoland

import (
	"fmt"
	"strconv"
	"strings"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	sdkparams "github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/store"
	tmver "github.com/gnolang/gno/tm2/pkg/version"
)

const (
	nodeParamHaltHeight     = "node:p:halt_height"
	nodeParamHaltMinVersion = "node:p:halt_min_version"

	// Valset keys live under the "valset" submodule of the node module.
	// Keep in sync with examples/gno.land/r/sys/params/valset.gno.
	//
	//   dirty    flag set by realm; EndBlocker clears after applying.
	//   proposed v3's full target valset.
	//   current  chain-managed: the set that becomes active at H+2 once
	//            the most recent EndBlock's updates apply. NOT the set
	//            actively signing the current block.
	//
	// Each "proposed"/"current" entry has the form
	// "<bech32-pubkey>:<decimal-power>".
	valsetDirtyPath    = "node:valset:dirty"
	valsetProposedPath = "node:valset:proposed"
	valsetCurrentPath  = "node:valset:current"

	// maxValsetEntries caps len(valset:proposed) at WillSetParam time.
	// v3 enforces 40 at proposal-creation; this is defense-in-depth at
	// 2.5x to protect against future writers that bypass v3's cap.
	maxValsetEntries = 100
)

// internalWriteCtxKey marks chain-internal writes (InitChainer,
// EndBlocker). User-routed writes (governance proposals via the
// generic params factories) carry no such value.
type internalWriteCtxKey struct{}

// nodeParamsKeeper implements a minimal ParamfulKeeper for the "node" module.
// It validates node-level parameters set through governance proposals.
type nodeParamsKeeper struct{}

// WillSetParam validates node parameters before they are written to the params store.
func (nodeParamsKeeper) WillSetParam(ctx sdk.Context, key string, value any) {
	switch key {
	case "p:halt_height":
		h, ok := value.(int64)
		if !ok {
			panic(fmt.Sprintf("halt_height must be an int64, got %T", value))
		}
		if h < 0 {
			panic(fmt.Sprintf("halt_height must be non-negative, got %d", h))
		}
		// Reject halt heights that are in the past or present.
		// h == 0 is the cancel sentinel and is always allowed.
		// safeBlockHeight handles genesis/test contexts where the block header may not be set.
		if curHeight := safeBlockHeight(ctx); h > 0 && curHeight > 0 && h <= curHeight {
			panic(fmt.Sprintf("halt_height %d must be greater than the current block height %d", h, curHeight))
		}
	case "p:halt_min_version":
		_, ok := value.(string)
		if !ok {
			panic(fmt.Sprintf("halt_min_version must be a string, got %T", value))
		}
	case "valset:dirty":
		// Just type-check; the bool value is opaque.
		// Note: dirty has no ctx-sentinel gate because the realm side
		// is gated by assertValsetCaller in r/sys/params/valset.gno;
		// dirty is bool-typed only and not safety-critical on its own.
		if _, ok := value.(bool); !ok {
			panic(fmt.Sprintf("valset:dirty must be a bool, got %T", value))
		}
	case "valset:proposed":
		// Validate each "<pubkey>:<power>" entry on write so a bad realm
		// can't seed garbage that EndBlocker has to recover from.
		entries, ok := value.([]string)
		if !ok {
			panic(fmt.Sprintf("valset:proposed must be []string, got %T", value))
		}
		if len(entries) > maxValsetEntries {
			panic(fmt.Sprintf("valset:proposed too long: %d > %d", len(entries), maxValsetEntries))
		}
		if _, err := abci.ParseValidatorUpdates(entries); err != nil {
			panic(fmt.Sprintf("invalid valset:proposed: %v", err))
		}
	case "valset:current":
		// Chain-only key. Use type-assertion idiom (codebase
		// convention; sdk/auth/params.go:204 etc.) rather than `!= true`,
		// which works but is non-idiomatic.
		v, _ := ctx.Value(internalWriteCtxKey{}).(bool)
		if !v {
			panic("valset:current is chain-managed; not writable via params")
		}
		entries, ok := value.([]string)
		if !ok {
			panic(fmt.Sprintf("valset:current must be []string, got %T", value))
		}
		if _, err := abci.ParseValidatorUpdates(entries); err != nil {
			panic(fmt.Sprintf("invalid valset:current (chain-internal corruption): %v", err))
		}
	default:
		if strings.HasPrefix(key, "p:") {
			panic(fmt.Sprintf("unknown node param key: %q", key))
		}
		if strings.HasPrefix(key, "valset:") {
			panic(fmt.Sprintf("unknown valset key: %q", key))
		}
	}
}

// checkNodeStartupParams reads halt-related params from the committed state and verifies:
//  1. The running binary meets the minimum version requirement set by governance.
//  2. A new (upgraded) binary is not started before the chain has actually halted.
//
// skipUpgradeHeight, if non-zero, skips all upgrade checks at that specific height.
func checkNodeStartupParams(prmk sdkparams.ParamsKeeperI, ms store.MultiStore, lastBlockHeight, skipUpgradeHeight int64) error {
	// Build a minimal read-only context with just the multistore and a placeholder chain ID.
	// We only need store access to read params; no block execution context is required.
	ctx := sdk.Context{}.WithMultiStore(ms).WithChainID("_")

	var haltHeight int64
	prmk.GetInt64(ctx, nodeParamHaltHeight, &haltHeight)

	var minVersion string
	prmk.GetString(ctx, nodeParamHaltMinVersion, &minVersion)

	// Nothing to check if no governance halt is configured.
	if haltHeight == 0 || minVersion == "" {
		return nil
	}

	// Allow skipping upgrade checks at a specific height (e.g., validator already migrated).
	if skipUpgradeHeight > 0 && skipUpgradeHeight == haltHeight {
		return nil
	}

	binaryVersion := tmver.Version

	// Check 1: Prevent old binaries from resuming after a halt.
	if lastBlockHeight >= haltHeight {
		if !meetsMinVersion(binaryVersion, minVersion) {
			return fmt.Errorf(
				"binary version %q does not meet the minimum version %q required by governance; "+
					"please upgrade to a compatible binary before restarting",
				binaryVersion, minVersion,
			)
		}
		return nil
	}

	// Check 2: Prevent new (upgraded) binaries from running before the halt height.
	// Any binary that meets the minimum version is rejected until the halt occurs.
	if meetsMinVersion(binaryVersion, minVersion) {
		return fmt.Errorf(
			"binary version %q is an upgrade intended for halt height %d, "+
				"but the chain is at height %d; please use the previous binary until the halt, "+
				"or set skip_upgrade_height = %d in config.toml if you have already migrated",
			binaryVersion, haltHeight, lastBlockHeight, haltHeight,
		)
	}

	return nil
}

// safeBlockHeight returns ctx.BlockHeight() or 0 if the context has no block header.
// This handles genesis and test contexts where the header may not be initialized.
func safeBlockHeight(ctx sdk.Context) (h int64) {
	defer func() { recover() }() //nolint:errcheck
	return ctx.BlockHeight()
}

// meetsMinVersion reports whether binaryVersion satisfies the minVersion requirement.
// Versions are expected to follow the "chain/gnolandX.Y" format used for gno.land chain releases.
// If either version cannot be parsed in that format, an exact string match is required.
func meetsMinVersion(binaryVersion, minVersion string) bool {
	if minVersion == "" {
		return true
	}

	bMajor, bMinor, bOK := parseGnolandVersion(binaryVersion)
	mMajor, mMinor, mOK := parseGnolandVersion(minVersion)

	if bOK && mOK {
		if bMajor != mMajor {
			return bMajor > mMajor
		}
		return bMinor >= mMinor
	}

	// Fall back to exact match if versions are not in the recognized format.
	return binaryVersion == minVersion
}

// parseGnolandVersion parses a version string like "chain/gnoland1.2" into its major and minor parts.
func parseGnolandVersion(v string) (major, minor int, ok bool) {
	const prefix = "chain/gnoland"
	if !strings.HasPrefix(v, prefix) {
		return 0, 0, false
	}
	rest := v[len(prefix):]
	dot := strings.IndexByte(rest, '.')
	if dot < 0 {
		return 0, 0, false
	}
	maj, err1 := strconv.Atoi(rest[:dot])
	mnr, err2 := strconv.Atoi(rest[dot+1:])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return maj, mnr, true
}
