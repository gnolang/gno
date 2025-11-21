package auth

import (
	"flag"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
)

type authRawFlags struct {
	auth *common.AuthFlags
	raw  bool
}

var defaultAuthRawFlags = &authRawFlags{
	raw: false,
}

func (f *authRawFlags) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&f.raw,
		"raw",
		defaultAuthRawFlags.raw,
		"output raw values, without descriptions",
	)
}
