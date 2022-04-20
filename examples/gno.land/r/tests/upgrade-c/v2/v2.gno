package v1

import "gno.land/r/tests/upgrade-c/root"

func Inc() {
	root.Inc()
}

func Get() int {
	return root.Get() * 2
}
