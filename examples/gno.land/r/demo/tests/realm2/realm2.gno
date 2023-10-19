package realm2

import (
	"gno.land/r/demo/tests/realm1"
)

func Do() {
	realm1.GetCounterPkg().Inc()
}
