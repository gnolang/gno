> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `p/nt/grants/v0` — grant lifecycle for realm grantees

The pure state machine behind [`r/gov/grants/v0`](../../../../r/gov/grants/v0): a `Program` of
`Grant`s whose grantees are realms. It holds no coins and reads no chain state. A desk derives the
grantee with `GranteeOf`, supplies the time and its balance, and performs the transfer `Claim`
hands back. Any grantor, a sub-DAO for instance, can run its own desk on it with its own `Rules`
and its own way of deciding; the GovDAO desk's pages and scripts stay in that realm.

```go
desk, err := grants.NewProgram(grants.Rules{Window: 14 * day, MinPeriod: 30 * day, Expiry: 60 * day, HoldFor: 14 * day, Advance: 20})

grantee, err := grants.GranteeOf(0, cur)                       // in a crossing function of the desk
id, err := desk.Apply(grantee, title, pitch, tranches, 30*day, time.Now())
err = desk.Approve(id, balance, sponsor, review, time.Now())    // reserves the total
err = desk.Report(grantee, id, "v3 is live", "/r/team/game", time.Now())
to, amount, err := desk.Claim(id, time.Now())                   // once due; the desk sends the coins
```

The package doc in `grants.gno` states the rules. Not to be confused with
[`gno.land/p/moul/grants/v0`](https://gno.land/p/moul/grants/v0), whose grantees are people and
whose members vote on each milestone.
