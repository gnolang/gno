package subscription

type RecurringSubscription struct {
	duration time.Duration
	amount   int64
	subs     *avl.Tree // std.Address -> time.Time
}

func (rs *RecurringSubscription) CheckOrigCaller() {
	send := std.GetOrigSend()
	caller := std.GetOrigCaller()

	// TODO implement
	good := false
	if !good {
		panic("you need to pay a subscription of " + rs.amount)
	}
}

// TODO: other subscription models
