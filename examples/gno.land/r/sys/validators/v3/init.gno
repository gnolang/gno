package validators

import "gno.land/p/nt/poa"

func init() {
	// The default valset protocol is PoA
	vp = poa.NewPoA()
}
