package params

import (
	"chain"

	"gno.land/r/gov/dao"
)

func NewSetFeeCollectorRequest(addr chain.Address) dao.ProposalRequest {
	return NewSysParamStringPropRequest(
		"auth", "p", "fee_collector",
		addr.String(),
	)
}
