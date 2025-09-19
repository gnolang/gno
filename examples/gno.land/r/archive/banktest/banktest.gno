package banktest

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"time"
)

type activity struct {
	caller   address
	sent     chain.Coins
	returned chain.Coins
	time     time.Time
}

func (act *activity) String() string {
	return act.caller.String() + " " +
		act.sent.String() + " sent, " +
		act.returned.String() + " returned, at " +
		act.time.Format("2006-01-02 3:04pm MST")
}

var latest [10]*activity

func addActivity(act *activity) {
	for i := len(latest) - 2; i >= 0; i-- {
		latest[i+1] = latest[i] // shift by +1.
	}
	latest[0] = act
}

// Deposit will take the coins (to the realm's pkgaddr) if returnAmount is 0,
// or return the specified return amount back to user.
func Deposit(cur realm, returnDenom string, returnAmount int64) string {
	runtime.AssertOriginCall()
	caller := runtime.OriginCaller()
	send := chain.Coins{{returnDenom, returnAmount}}
	// record activity
	act := &activity{
		caller:   caller,
		sent:     banker.OriginSend(),
		returned: send,
		time:     time.Now(),
	}
	addActivity(act)
	// return if any.
	if returnAmount > 0 {
		banker_ := banker.NewBanker(banker.BankerTypeOriginSend)
		pkgaddr := runtime.CurrentRealm().Address()
		// TODO: use std.Coins constructors, this isn't generally safe.
		banker_.SendCoins(pkgaddr, caller, send)
		return "returned!"
	} else {
		return "thank you!"
	}
}

func bankerAddr(cur realm) address {
	return runtime.CurrentRealm().Address()
}

func Render(path string) string {
	// get realm coins.
	banker_ := banker.NewBanker(banker.BankerTypeReadonly)
	coins := banker_.GetCoins(bankerAddr(cross))

	// render
	res := ""
	res += "## recent activity\n"
	res += "\n"
	for _, act := range latest {
		if act == nil {
			break
		}
		res += " * " + act.String() + "\n"
	}
	res += "\n"
	res += "## total deposits\n"
	res += coins.String()
	return res
}
