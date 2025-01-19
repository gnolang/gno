package gnokey

import (
	"flag"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
)

type gnokeyFlags struct {
	home                  string
	insecurePasswordStdin bool
}

var defaultGnokeyFlags = &gnokeyFlags{
	home:                  gnoenv.HomeDir(),
	insecurePasswordStdin: false,
}

func (f *gnokeyFlags) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&f.home,
		"home",
		defaultGnokeyFlags.home,
		"gnokey home directory",
	)

	fs.BoolVar(
		&f.insecurePasswordStdin,
		"insecure-password-stdin",
		defaultGnokeyFlags.insecurePasswordStdin,
		"WARNING! take password from stdin",
	)
}
