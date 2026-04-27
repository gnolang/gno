package gnoland

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	sdkparams "github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/store"
	tmver "github.com/gnolang/gno/tm2/pkg/version"
)

const (
	nodeParamHaltHeight     = "node:p:halt_height"
	nodeParamHaltMinVersion = "node:p:halt_min_version"
)

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
	default:
		if strings.HasPrefix(key, "p:") {
			panic(fmt.Sprintf("unknown node param key: %q", key))
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
