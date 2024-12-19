package vm

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
)

const (
	sysUsersPkgParamPath = "gno.land/r/sys/params.sys.users_pkgpath.string"
	chainDomainParamPath = "gno.land/r/sys/params.chain_domain.string"
)

func (vm *VMKeeper) getChainDomainParam(ctx sdk.Context) (string, error) {
	chainDomain, err := vm.prmk.GetString(ctx, chainDomainParamPath)
	if errors.Is(err, params.ErrMissingParamValue) || chainDomain == "" {
		// Return default
		return "gno.land", nil
	}

	if err != nil {
		return "", fmt.Errorf("unable to load param %s, %w", sysUsersPkgParamPath, err)
	}

	return chainDomain, nil
}
