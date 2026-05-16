package namereg

import (
	"chain"
	"chain/runtime"

	susers "gno.land/r/sys/users"
)

func init(cur realm) {
	if runtime.ChainHeight() == 0 {
		susers.AddControllerAtGenesis(cross2(cur), chain.PackageAddress("gno.land/r/sys/namereg/v1"))
	}
}
