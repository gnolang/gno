package predictionmarket

import (
	"strconv"
	"strings"

	"gno.land/p/nt/avl/v0"

	"chain"
	"chain/banker"
	"chain/runtime"
	"chain/runtime/unsafe"
)

const (
	OutcomeHome = 0
	OutcomeAway = 1
	OutcomeDraw = 2

	betDenom = "ugnot"
)

type Bet struct {
	bettor  address
	outcome int
	amount  int64
}

type Market struct {
	id          int
	description string
	resolved    bool
	winner      int
	pools       [3]int64
	bets        []*Bet
}

var (
	markets avl.Tree
	nextID  int
	admin   address
)

func init() {
	admin = unsafe.OriginCaller()
	nextID = 1
}

func CreateMarket(cur realm, description string) int {
	caller := cur.Previous().Address()
	if caller != admin {
		panic("only admin can create markets")
	}
	id := nextID
	nextID++
	m := &Market{
		id:          id,
		description: description,
		winner:      -1,
	}
	markets.Set(strconv.Itoa(id), m)
	return id
}

func PlaceBet(cur realm, marketID int, outcome int) {
	runtime.AssertOriginCall()
	if !cur.Previous().IsUserCall() {
		panic("only direct EOA calls can place bets")
	}
	if outcome < 0 || outcome > 2 {
		panic("invalid outcome")
	}

	raw, exists := markets.Get(strconv.Itoa(marketID))
	if !exists {
		panic("market does not exist")
	}
	m := raw.(*Market)
	if m.resolved {
		panic("market already resolved")
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

	caller := cur.Previous().Address()
	m.pools[outcome] += amount
	m.bets = append(m.bets, &Bet{bettor: caller, outcome: outcome, amount: amount})
}

func ResolveMarket(cur realm, marketID int, winningOutcome int) {
	caller := cur.Previous().Address()
	if caller != admin {
		panic("only admin can resolve markets")
	}
	if winningOutcome < 0 || winningOutcome > 2 {
		panic("invalid outcome")
	}
	raw, exists := markets.Get(strconv.Itoa(marketID))
	if !exists {
		panic("market does not exist")
	}
	m := raw.(*Market)
	if m.resolved {
		panic("market already resolved")
	}
	m.resolved = true
	m.winner = winningOutcome
}

func ClaimWinnings(cur realm, marketID int) {
	caller := cur.Previous().Address()
	raw, exists := markets.Get(strconv.Itoa(marketID))
	if !exists {
		panic("market does not exist")
	}
	m := raw.(*Market)
	if !m.resolved {
		panic("market not resolved yet")
	}

	winningPool := m.pools[m.winner]
	if winningPool == 0 {
		panic("no winning pool")
	}

	var totalPool int64
	for _, p := range m.pools {
		totalPool += p
	}

	var payout int64
	for _, b := range m.bets {
		if b.bettor == caller && b.outcome == m.winner && b.amount > 0 {
			payout += (b.amount * totalPool) / winningPool
			b.amount = 0
		}
	}
	if payout == 0 {
		panic("nothing to claim")
	}

	bkr := banker.NewBanker(banker.BankerTypeRealmSend, cur)
	bkr.SendCoins(cur.Address(), caller, chain.Coins{{Denom: betDenom, Amount: payout}})
}

func Render(path string) string {
	var sb strings.Builder
	sb.WriteString("Prediction Market Realm\nAdmin: " + admin.String() + "\n\n")
	markets.Iterate("", "", func(key string, value interface{}) bool {
		m := value.(*Market)
		status := "open"
		if m.resolved {
			status = "resolved, winner: " + strconv.Itoa(m.winner)
		}
		sb.WriteString("Market #" + strconv.Itoa(m.id) + ": " + m.description +
			" | Home: " + strconv.FormatInt(m.pools[OutcomeHome], 10) +
			" Away: " + strconv.FormatInt(m.pools[OutcomeAway], 10) +
			" Draw: " + strconv.FormatInt(m.pools[OutcomeDraw], 10) +
			" [" + status + "]\n")
		return false
	})
	return sb.String()
}
