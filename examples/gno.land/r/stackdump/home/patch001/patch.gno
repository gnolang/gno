package patch


import (
    index "gno.land/r/stackdump/home"
)

func init() {
    index.Set("foo", "hello world")
}