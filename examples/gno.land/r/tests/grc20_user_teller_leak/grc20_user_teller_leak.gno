// Package grc20_user_teller_leak is a regression fixture for leaked tellers.
package grc20_user_teller_leak

type teller interface {
	Transfer(_ int, rlm realm, to address, amount int64) error
}

var leaked teller

func Store(cur realm, t teller) {
	leaked = t
}

func Transfer(cur realm, to address, amount int64) error {
	return leaked.Transfer(0, cur, to, amount)
}
