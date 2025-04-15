package registree

import "gno.land/r/leon/registry"

func ManualRegister() {
	fn := registry.Register
	crossing(fn)
}

//func init() {
//	registry.Register()
//}
