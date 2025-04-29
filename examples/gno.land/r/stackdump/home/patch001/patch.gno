package patch

import (
	index "gno.land/r/stackdump/home"
)

func init() { // XXX able to mutate home data - is this by design?
	index.Set("foo", "hello world")
}

/*

# Confirm that crossing is not allowed in this case via API

gnokey maketx call -pkgpath "gno.land/r/stackdump/home" -func "SetString" -args "foo" -args "Bar" \
-gas-fee 1000000ugnot -gas-wanted 5000000 -broadcast -chainid "dev" -remote "tcp://127.0.0.1:26657" ADDRESS

# Correctly produces the error

Data: missing crossing() after cross call to SetString from  to gno.land/r/stackdump/home

*/
