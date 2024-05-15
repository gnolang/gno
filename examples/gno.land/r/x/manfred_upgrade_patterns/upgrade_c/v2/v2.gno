package v1

import "gno.land/r/x/manfred_upgrade_patterns/upgrade_c/root"

func Inc() {
	root.Inc()
}

func Get() int {
	return root.Get() * 2
}
