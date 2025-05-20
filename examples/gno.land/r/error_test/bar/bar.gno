// PATH: bar/a.gno
package bar

import (
 "gno.land/r/error_test/foo"
)

func Bar() int64 {
    return foo.Foo()
}