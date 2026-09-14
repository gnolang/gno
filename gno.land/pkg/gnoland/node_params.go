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
	//   proposed v0's full target valset.
	//   current  chain-managed: the set that becomes active at H+2 once
	//            the most recent EndBlock's updates apply. NOT the set
	//            actively signing the current block.
	//
	// Each "proposed"/"current" entry has the form
	// "<bech32-pubkey>:<decimal-power>".
	valsetDirtyPath    = "node:valset:dirty"
	valsetProposedPath = "node:valset:proposed"
	valsetCurrentPath  = "node:valset:current"

	// pubkey_types: chain-managed mirror of the consensus-params validator allow-list, so realms can read it.
	valsetPubKeyTypesPath = "node:valset:pubkey_types"

	// maxValsetEntries caps len(valset:proposed) at WillSetParam time.
	// v0 enforces 40 at proposal-creation; this is defense-in-depth at
	// 2.5x to protect against future writers that bypass v0's cap.
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
	case "valset:pubkey_types":
		// Chain-only mirror; reject user-routed writes (like valset:current).
		v, _ := ctx.Value(internalWriteCtxKey{}).(bool)
		if !v {
			panic("valset:pubkey_types is chain-managed; not writable via params")
		}
		if _, ok := value.([]string); !ok {
			panic(fmt.Sprintf("valset:pubkey_types must be []string, got %T", value))
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
//
// Two release-tag shapes are understood, ordered against each other as a single
// line (see RELEASING.md):
//
//	vMAJOR.MINOR.PATCH        the current shape, e.g. "v1.2.0"
//	chain/gnolandMAJOR.MINOR  betanet's retired shape, e.g. "chain/gnoland1.1"
//
// Anything else does not parse: "develop" (a plain `go build`),
// "master.3335+bc43a5fb7" (an off-tag `make` build), or an un-numbered chain tag
// such as "chain/mainnet". A binary whose version does not parse satisfies no
// floor, which is the intended outcome — an ad-hoc build must not pass an
// upgrade gate.
//
// A minVersion that does not parse is the dangerous case, and the reason this
// function takes shapes rather than one: it degrades to byte equality, so the
// correctly-upgraded binary is refused alongside the stale ones and the chain
// cannot restart at all. Governance must name a parseable release tag; the
// release tooling in misc/release refuses to emit a proposal that does not.
func meetsMinVersion(binaryVersion, minVersion string) bool {
	if minVersion == "" {
		return true
	}

	bv, bOK := parseReleaseVersion(binaryVersion)
	mv, mOK := parseReleaseVersion(minVersion)

	if bOK && mOK {
		return bv.compare(mv) >= 0
	}

	// Fall back to exact match if versions are not in the recognized format.
	return binaryVersion == minVersion
}

// releaseVersion is a parsed gno.land release tag.
type releaseVersion struct {
	major, minor, patch int
	// pre is the semver pre-release suffix without its leading '-', empty for a
	// final release. A pre-release sorts below the release it leads to, so
	// "v1.3.0-rc.1" does not satisfy a "v1.3.0" floor.
	pre string
}

// compare returns -1, 0 or +1 as v sorts before, equal to, or after o.
func (v releaseVersion) compare(o releaseVersion) int {
	for _, pair := range [][2]int{
		{v.major, o.major},
		{v.minor, o.minor},
		{v.patch, o.patch},
	} {
		if pair[0] != pair[1] {
			if pair[0] < pair[1] {
				return -1
			}
			return 1
		}
	}

	// Equal numbers: a release outranks any pre-release of itself. Two
	// pre-releases are ordered by plain string comparison, which matches semver
	// for the "rc.N" shape we use and is only approximate beyond it.
	switch {
	case v.pre == o.pre:
		return 0
	case v.pre == "":
		return 1
	case o.pre == "":
		return -1
	case v.pre < o.pre:
		return -1
	default:
		return 1
	}
}

// legacyChainPrefix is betanet's tag shape. It is frozen: chain/gnoland1.0 and
// chain/gnoland1.1 are the only two tags that ever used it, and they are the
// version strings compiled into the binaries that ran that chain.
const legacyChainPrefix = "chain/gnoland"

// parseReleaseVersion parses a release tag in either supported shape.
func parseReleaseVersion(v string) (releaseVersion, bool) {
	if rest, ok := strings.CutPrefix(v, legacyChainPrefix); ok {
		// "chain/gnolandMAJOR.MINOR" — no patch component.
		majorStr, minorStr, found := strings.Cut(rest, ".")
		if !found {
			return releaseVersion{}, false
		}
		major, err1 := strconv.Atoi(majorStr)
		minor, err2 := strconv.Atoi(minorStr)
		if err1 != nil || err2 != nil {
			return releaseVersion{}, false
		}
		return releaseVersion{major: major, minor: minor}, true
	}

	rest, ok := strings.CutPrefix(v, "v")
	if !ok {
		return releaseVersion{}, false
	}

	// Drop the build metadata; semver says it takes no part in ordering.
	rest, _, _ = strings.Cut(rest, "+")
	rest, pre, _ := strings.Cut(rest, "-")

	parts := strings.Split(rest, ".")
	if len(parts) != 3 {
		return releaseVersion{}, false
	}
	nums := make([]int, 3)
	for i, p := range parts {
		// Reject "+1", "-1" and leading zeros, which Atoi would otherwise accept
		// or silently normalise into a tag that is not the one that was pushed.
		if p == "" || (len(p) > 1 && p[0] == '0') {
			return releaseVersion{}, false
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return releaseVersion{}, false
		}
		nums[i] = n
	}
	return releaseVersion{major: nums[0], minor: nums[1], patch: nums[2], pre: pre}, true
}
