package patch

import (
	index "gno.land/r/stackdump/home"
)

func Bug() {
	crossing()
	index.Set("foo", "hello world")
}
