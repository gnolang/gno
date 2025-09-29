package params

import (
	"gno.land/r/gov/dao"
)

func NewSetFeeCollectorRequest(addr address) dao.ProposalRequest {
	return NewSysParamStringPropRequest(
		"auth", "p", "fee_collector",
		addr.String(),
	)
}
