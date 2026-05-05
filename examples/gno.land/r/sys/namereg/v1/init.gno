package namereg

import (
	"chain"
	"chain/runtime"

	susers "gno.land/r/sys/users"
)

func init() {
	if runtime.ChainHeight() == 0 {
		susers.AddControllerAtGenesis(cross, chain.PackageAddress("gno.land/r/sys/namereg/v1"))
	}
}
