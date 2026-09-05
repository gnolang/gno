package predictionmarket

import (
	"math/overflow"
	"strconv"
	"strings"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ownable/v0"

	"chain"
	"chain/banker"
	"chain/runtime/unsafe"
)

const (
	OutcomeHome = 0
	OutcomeAway = 1
	OutcomeDraw = 2

	betDenom = "ugnot"
)

// Market is a single prediction market. The admin (owner) alone
// decides when to resolve it and which outcome won -- there is no
// oracle, dispute window, or betting cutoff. Bettors are trusting the
// admin completely; this is documented, not hidden, and the admin is
// blocked from placing bets in their own markets (see PlaceBet) to
// remove the most obvious form of self-dealing, but nothing prevents
// an admin from simply resolving in whatever direction benefits them.
type Market struct {
	id          int
	description string
	allowDraw   bool
	resolved    bool
	winner      int
	noWinners   bool // true if the resolved outcome received no bets; triggers a full refund instead of a proportional payout
	pools       [3]int64
	bets        []*Bet
}

type Bet struct {
	bettor  address
	outcome int
	amount  int64
}

var (
	markets avl.Tree
	nextID  int
	owner   *ownable.Ownable
)

func init() {
	owner = ownable.NewWithAddress(unsafe.OriginCaller())
	nextID = 1
}

// TransferOwnership hands off admin control of every market (past and
// future) to newOwner.
func TransferOwnership(cur realm, newOwner address) {
	if err := owner.TransferOwnership(0, cur, newOwner); err != nil {
		panic(err.Error())
	}
}

// marketKey zero-pads the id so markets.Iterate (which walks keys as
// strings) lists markets in numeric order instead of lexical order
// (which would otherwise put market 10 before market 2).
func marketKey(id int) string {
	s := strconv.Itoa(id)
	for len(s) < 10 {
		s = "0" + s
	}
	return s
}

// CreateMarket opens a new market. allowDraw controls whether
// OutcomeDraw is a valid bet for this specific market -- set it to
// false for contests that cannot end in a draw.
func CreateMarket(cur realm, description string, allowDraw bool) int {
	caller := cur.Previous().Address()
	owner.AssertOwnedBy(caller)

	id := nextID
	nextID++
	m := &Market{
		id:          id,
		description: description,
		allowDraw:   allowDraw,
		winner:      -1,
	}
	markets.Set(marketKey(id), m)
	return id
}

// PlaceBet attaches ugnot (read via unsafe.OriginSend) to a bet on
// outcome. cur.Previous().IsUserCall() rejects both a direct call from
// another realm and a gnokey maketx run ephemeral script (which has a
// non-empty, but not user, pkgPath) -- only a plain EOA transaction
// gets through. The admin is additionally blocked from betting in any
// market: they alone decide the resolution, so allowing them to also
// hold a position would let them resolve in their own favor risk-free.
func PlaceBet(cur realm, marketID int, outcome int) {
	if !cur.Previous().IsUserCall() {
		panic("only direct EOA calls can place bets")
	}

	caller := cur.Previous().Address()
	if owner.OwnedBy(caller) {
		panic("admin cannot place bets")
	}

	raw := markets.Get(marketKey(marketID))
	if raw == nil {
		panic("market does not exist")
	}
	m := raw.(*Market)
	if m.resolved {
		panic("market already resolved")
	}
	if outcome < OutcomeHome || outcome > OutcomeAway && !(outcome == OutcomeDraw && m.allowDraw) {
		panic("invalid outcome for this market")
	}

	sent := unsafe.OriginSend()
	var amount int64
	for _, c := range sent {
		if c.Denom == betDenom {
			amount = c.Amount
		}
	}
	if amount <= 0 {
		panic("must attach ugnot to place a bet")
	}

	m.pools[outcome] = overflow.Add64p(m.pools[outcome], amount)
	m.bets = append(m.bets, &Bet{bettor: caller, outcome: outcome, amount: amount})
}

// ResolveMarket locks in winningOutcome. If nobody bet on that
// outcome, the market is marked noWinners and ClaimWinnings refunds
// every bettor their original stake instead of computing a
// proportional payout against an empty winning pool -- otherwise the
// escrow would be stuck in the realm forever, since there would be no
// way to divide zero into a payout.
func ResolveMarket(cur realm, marketID int, winningOutcome int) {
	caller := cur.Previous().Address()
	owner.AssertOwnedBy(caller)

	raw := markets.Get(marketKey(marketID))
	if raw == nil {
		panic("market does not exist")
	}
	m := raw.(*Market)
	if m.resolved {
		panic("market already resolved")
	}
	if winningOutcome < OutcomeHome || winningOutcome > OutcomeAway && !(winningOutcome == OutcomeDraw && m.allowDraw) {
		panic("invalid outcome for this market")
	}

	m.resolved = true
	m.winner = winningOutcome
	if m.pools[winningOutcome] == 0 {
		m.noWinners = true
	}
}

// ClaimWinnings pays out the caller's share of a resolved market. For
// a normal resolution this is proportional to the caller's stake in
// the winning pool, drawn from the full pool. For a market marked
// noWinners (nobody backed the resolved outcome), this instead refunds
// every one of the caller's bets in full, regardless of which outcome
// they were on -- a push, not a proportional payout.
func ClaimWinnings(cur realm, marketID int) {
	caller := cur.Previous().Address()
	raw := markets.Get(marketKey(marketID))
	if raw == nil {
		panic("market does not exist")
	}
	m := raw.(*Market)
	if !m.resolved {
		panic("market not resolved yet")
	}

	var payout int64
	if m.noWinners {
		for _, b := range m.bets {
			if b.bettor == caller && b.amount > 0 {
				payout = overflow.Add64p(payout, b.amount)
				b.amount = 0
			}
		}
	} else {
		winningPool := m.pools[m.winner]
		var totalPool int64
		for _, p := range m.pools {
			totalPool = overflow.Add64p(totalPool, p)
		}
		for _, b := range m.bets {
			if b.bettor == caller && b.outcome == m.winner && b.amount > 0 {
				share := overflow.Mul64p(b.amount, totalPool) / winningPool
				payout = overflow.Add64p(payout, share)
				b.amount = 0
			}
		}
	}
	if payout == 0 {
		panic("nothing to claim")
	}

	bkr := banker.NewBanker(banker.BankerTypeRealmSend, cur)
	bkr.SendCoins(cur.Address(), caller, chain.Coins{{Denom: betDenom, Amount: payout}})
}

// Render shows every open/resolved market, or a single market's detail
// when called with ?id=<marketID>.
func Render(path string) string {
	if id, ok := parseIDQuery(path); ok {
		return renderMarket(id)
	}

	var sb strings.Builder
	sb.WriteString("# Prediction Market Realm\n\n")
	sb.WriteString("Admin: " + owner.Owner().String() + "\n\n")
	markets.Iterate("", "", func(key string, value interface{}) bool {
		m := value.(*Market)
		sb.WriteString(marketSummaryLine(m))
		return false
	})
	return sb.String()
}

func renderMarket(id int) string {
	raw := markets.Get(marketKey(id))
	if raw == nil {
		return "Market not found.\n"
	}
	m := raw.(*Market)

	var sb strings.Builder
	sb.WriteString("# Market #" + strconv.Itoa(m.id) + "\n\n")
	sb.WriteString(m.description + "\n\n")
	sb.WriteString(marketSummaryLine(m))
	return sb.String()
}

func marketSummaryLine(m *Market) string {
	status := "open"
	if m.resolved {
		if m.noWinners {
			status = "resolved, no winners (refund)"
		} else {
			status = "resolved, winner: " + strconv.Itoa(m.winner)
		}
	}
	line := "Market #" + strconv.Itoa(m.id) + ": " + m.description +
		" | Home: " + strconv.FormatInt(m.pools[OutcomeHome], 10) +
		" Away: " + strconv.FormatInt(m.pools[OutcomeAway], 10)
	if m.allowDraw {
		line += " Draw: " + strconv.FormatInt(m.pools[OutcomeDraw], 10)
	}
	line += " [" + status + "]\n"
	return line
}

// parseIDQuery extracts ?id=<n> from path, if present and valid.
func parseIDQuery(path string) (int, bool) {
	const prefix = "?id="
	idx := strings.Index(path, prefix)
	if idx == -1 {
		return 0, false
	}
	rest := path[idx+len(prefix):]
	if amp := strings.Index(rest, "&"); amp != -1 {
		rest = rest[:amp]
	}
	id, err := strconv.Atoi(rest)
	if err != nil {
		return 0, false
	}
	return id, true
}
