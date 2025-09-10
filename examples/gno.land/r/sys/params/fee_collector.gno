package params

import (
	"std"

	"gno.land/r/gov/dao"
)

func NewSetFeeCollectorRequest(addr std.Address) dao.ProposalRequest {
	return NewSysParamStringPropRequest(
		"auth", "p", "fee_collector",
		addr.String(),
	)
}
