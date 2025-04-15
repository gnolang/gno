package registree

import registry "gno.land/r/leon/registry"

func ManualRegister() {
	crossing(registry.Register)
}

func init() {
	crossing(registry.Register)
}
