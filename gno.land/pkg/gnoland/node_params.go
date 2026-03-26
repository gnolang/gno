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
func (nodeParamsKeeper) WillSetParam(_ sdk.Context, key string, value any) {
	switch key {
	case "p:halt_height":
		h, ok := value.(int64)
		if !ok {
			panic(fmt.Sprintf("halt_height must be an int64, got %T", value))
		}
		if h < 0 {
			panic(fmt.Sprintf("halt_height must be non-negative, got %d", h))
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

// checkNodeStartupParams reads halt-related params from the committed state and verifies
// that the running binary satisfies any minimum version requirement set by governance.
// This prevents an old binary from accidentally resuming a chain that was halted for an upgrade.
func checkNodeStartupParams(prmk sdkparams.ParamsKeeperI, ms store.MultiStore) error {
	// Build a minimal read-only context with just the multistore and a placeholder chain ID.
	// We only need store access to read params; no block execution context is required.
	ctx := sdk.Context{}.WithMultiStore(ms).WithChainID("_")

	var minVersion string
	prmk.GetString(ctx, nodeParamHaltMinVersion, &minVersion)
	if minVersion == "" {
		return nil
	}

	binaryVersion := tmver.Version
	if !meetsMinVersion(binaryVersion, minVersion) {
		return fmt.Errorf(
			"binary version %q does not meet the minimum version %q required by governance; "+
				"please upgrade to a compatible binary before restarting",
			binaryVersion, minVersion,
		)
	}

	return nil
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
	min, err2 := strconv.Atoi(rest[dot+1:])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return maj, min, true
}
