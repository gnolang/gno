package auth

import (
	"flag"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
)

type authRawFlags struct {
	auth *common.AuthFlags
	raw  bool
}

func (f *authRawFlags) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&f.raw,
		"raw",
		false,
		"output raw values, without descriptions",
	)
}
