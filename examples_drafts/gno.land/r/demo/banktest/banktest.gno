package banktest

import (
	"std"
	"time"
)

type activity struct {
	caller   std.Address
	sent     std.Coins
	returned std.Coins
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
	std.AssertOriginCall()
	caller := std.OriginCaller()
	send := std.Coins{{returnDenom, returnAmount}}
	// record activity
	act := &activity{
		caller:   caller,
		sent:     std.OriginSend(),
		returned: send,
		time:     time.Now(),
	}
	addActivity(act)
	// return if any.
	if returnAmount > 0 {
		banker := std.NewBanker(std.BankerTypeOriginSend)
		pkgaddr := std.CurrentRealm().Address()
		// TODO: use std.Coins constructors, this isn't generally safe.
		banker.SendCoins(pkgaddr, caller, send)
		return "returned!"
	} else {
		return "thank you!"
	}
}

func bankerAddr(cur realm) std.Address {
	return std.CurrentRealm().Address()
}

func Render(path string) string {
	// get realm coins.
	banker := std.NewBanker(std.BankerTypeReadonly)
	coins := banker.GetCoins(bankerAddr(cross))

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
