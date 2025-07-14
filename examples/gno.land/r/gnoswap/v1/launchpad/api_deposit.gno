package launchpad

// ApiGetDepositByDepositId retrieves deposit information by deposit ID.
func ApiGetDepositByDepositId(depositId string) string {
	deposit, exist := deposits.Get(depositId)
	if !exist {
		return ""
	}

	builder := metaBuilder()
	depositBuilder(builder, deposit.(*Deposit))

	return marshal(builder.Node())
}
