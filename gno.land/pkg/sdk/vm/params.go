package vm

import "github.com/gnolang/gno/tm2/pkg/sdk"

const (
	sysUsersPkgParamPath = "gno.land/r/sys/params.sys.users_pkgpath.string"
	chainDomainParamPath = "gno.land/r/sys/params.chain.domain.string"
)

func (vm *VMKeeper) getChainDomainParam(ctx sdk.Context) string {
	chainDomain := "gno.land" // default
	vm.prmk.GetString(ctx, chainDomainParamPath, &chainDomain)
	return chainDomain
}

func (vm *VMKeeper) getSysUsersPkgParam(ctx sdk.Context) string {
	var sysUsersPkg string
	vm.prmk.GetString(ctx, sysUsersPkgParamPath, &sysUsersPkg)
	return sysUsersPkg
}
