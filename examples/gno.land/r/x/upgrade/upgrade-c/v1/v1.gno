package v1

import "gno.land/r/x/upgrade/upgrade-c/root"

func Inc() {
	root.Inc()
}

func Get() int {
	return root.Get() * 1
}
