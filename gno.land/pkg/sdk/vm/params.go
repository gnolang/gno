package vm

import "github.com/gnolang/gno/tm2/pkg/sdk"

const (
	chainTzParamPath = "gno.land/r/sys/params.chain_tz.string"
)

func (vm *VMKeeper) getChainTzParam(ctx sdk.Context) string {
	chainTz := "UTC" // default
	vm.prmk.GetString(ctx, chainTzParamPath, &chainTz)
	return chainTz
}
