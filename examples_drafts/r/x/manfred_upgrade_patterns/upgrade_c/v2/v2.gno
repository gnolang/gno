package v2

import "gno.land/r/x/manfred_upgrade_patterns/upgrade_c/root"

func Inc() int {
	return root.Inc(1000)
}

func Get() int {
	return root.Get() * 2
}
