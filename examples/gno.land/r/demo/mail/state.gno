package mail

import (
	"std"

	"gno.land/p/demo/avl"
)

// realm state
var (
	addr2Mailbox             = avl.Tree{}                                 // Address -> *Mailbox
	counter      int         = 0                                          // number of mails ever sent
	admin        std.Address = "g1fjh9y7ausp27dqsdq0qrcsnmgvwm6829v2au7d" // @grepsuzette

	Fee std.Coin = std.Coin{
		Denom:  "ugnot",
		Amount: 250 * 1000, // should never cost more $0.12 or 0.25
	}
)
