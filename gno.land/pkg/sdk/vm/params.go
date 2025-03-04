package vm

import "github.com/gnolang/gno/tm2/pkg/sdk"

const (
	sysNamesPkgParamPath = "gno.land/r/sys/params.sys.names_pkgpath.string"
	chainDomainParamPath = "gno.land/r/sys/params.chain_domain.string"
)

func (vm *VMKeeper) getChainDomainParam(ctx sdk.Context) string {
	chainDomain := "gno.land" // default
	vm.prmk.GetString(ctx, chainDomainParamPath, &chainDomain)
	return chainDomain
}

func (vm *VMKeeper) getSysNamesPkgParam(ctx sdk.Context) string {
	var sysNamesPkg string
	vm.prmk.GetString(ctx, sysNamesPkgParamPath, &sysNamesPkg)
	return sysNamesPkg
}
