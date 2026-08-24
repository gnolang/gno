package params

import (
	"gno.land/r/gov/dao"
)

func NewSetFeeCollectorRequest(cur realm, addr address) dao.ProposalRequest {
	return NewSysParamStringPropRequest(cur,
		"auth", "p", "fee_collector",
		addr.String(),
	)
}
