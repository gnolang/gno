// Package subscription provides recurring and lifetime subscription
// primitives. Payment is pulled through a caller-supplied
// *allowancesender.AllowanceSender (a cap-token whose construction is
// gated on banker.IsCanonical, so the lib can rely on Send actually
// moving real coins). Subscriber identity is decoupled from cap-token
// authority; the lib does not inspect call-chain shape.
//
// Each subscription instance locks its payment destination at
// construction time (payeeAddr) and pulls a fixed per-period amount
// from the cap on each Subscribe call. A leaked struct pointer cannot
// redirect payment.
//
// Recurring Subscription:
//
//	// In the consumer realm, at deployment:
//	var recSub = recurring.NewRecurringSubscription(
//	    time.Hour*24*30,                    // 30-day periods
//	    100,                                // 100 ugnot per period
//	    runtime.CurrentRealm().Address(),   // payee = this realm
//	)
//
//	// Wrapper exposed to users. The wrapper builds an AllowanceSender
//	// from a real banker, defers Close, and forwards the call. Subscribe
//	// pulls 100 ugnot from the consumer realm (whose -send envelope
//	// the wrapper validated) and records `subscriber` as active until
//	// time.Now() + 30 days.
//	func Subscribe(_ realm, subscriber address) {
//	    if !runtime.PreviousRealm().IsUserCall() { panic("EOA only") }
//	    src := banker.NewBanker(banker.BankerTypeRealmSend)
//	    al := allowancesender.New(src, runtime.CurrentRealm().Address(),
//	        chain.NewCoins(chain.NewCoin("ugnot", recSub.GetAmount())))
//	    defer al.Close()
//	    if err := recSub.Subscribe(al, subscriber); err != nil {
//	        panic(err)
//	    }
//	}
//
// Lifetime Subscription:
//
//	var lifeSub = lifetime.NewLifetimeSubscription(
//	    500,                                // 500 ugnot one-time
//	    runtime.CurrentRealm().Address(),
//	)
//
//	// Same wrapper shape as recurring.
//
// Both libraries trust whatever address the consumer's wrapper passes
// as `subscriber` — the wrapper's policy decides who's allowed to
// register, who can be gifted, etc. Self-Subscribe and gift collapse
// into the same call with different `subscriber` arguments.
package subscription
