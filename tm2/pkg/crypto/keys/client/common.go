package client

import (
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type BaseOptions struct {
	Home                  string
	Remote                string
	Quiet                 bool
	InsecurePasswordStdin bool
	Config                string
	// OnTxSuccess is called when the transaction tx succeeds. It can, for example,
	// print info in the result. If OnTxSuccess is nil, print basic info.
	OnTxSuccess func(tx std.Tx, res *ctypes.ResultBroadcastTxCommit)
}

var DefaultBaseOptions = BaseOptions{
	Home:                  "",
	Remote:                "127.0.0.1:26657",
	Quiet:                 false,
	InsecurePasswordStdin: false,
	Config:                "",
}
