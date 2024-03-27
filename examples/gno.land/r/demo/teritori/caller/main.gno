package caller

import (
    "gno.land/r/demo/teritori/callee"
)

func init() {
    callee.SetCaller()
}