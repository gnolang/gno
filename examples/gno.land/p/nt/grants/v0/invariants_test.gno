package grants

import (
	"strconv"
	"testing"
	"time"
)

// TestInvariantsUnderRandomOperations drives a program through a long,
// deterministic sequence of random operations from several grantees,
// including sub-realm identities, with a simulated pool, and checks after
// every step what must always hold:
//
//   - the pool never holds less than what is committed;
//   - committed is exactly what active grants still have to receive;
//   - the active and waiting indexes hold exactly the open grants, oldest first;
//   - a grantee never has two pending requests;
//   - a closed grant never changes status again;
//   - nothing is paid before it is due, and each payment is the next tranche.
func TestInvariantsUnderRandomOperations(t *testing.T) {
	for _, seed := range []uint64{42, 7, 1234567, 99} {
		randomRun(t, seed)
	}
}

func randomRun(t *testing.T, seed uint64) {
	p := program(t)
	grantees := []string{golf, golf + "#dao1", golf + "#dao2", other, "gno.land/r/test/golf/v2"}
	var (
		balance int64 = 1000
		now           = t0
		final         = map[int]string{}
	)
	rnd := func(n int) int {
		seed = seed*6364136223846793005 + 1442695040888963407
		return int((seed >> 33) % uint64(n))
	}

	for step := 0; step < 400; step++ {
		now = now.Add(time.Duration(rnd(10*24)) * time.Hour)
		id := rnd(p.Size()+2) + 1
		before, _ := p.Get(id)
		switch rnd(10) {
		case 0, 1:
			n := rnd(3) + 1
			ts := make([]Tranche, n)
			for i := range ts {
				ts[i] = Tranche{"d" + strconv.Itoa(i), int64(rnd(200) + 1)}
			}
			p.Apply(grantees[rnd(len(grantees))], "t", "p", ts, month+time.Duration(rnd(60))*day, now)
		case 2:
			balance += int64(rnd(300))
		case 3:
			p.Approve(id, balance, "", "", now)
		case 4:
			if g, ok := p.Get(id); ok {
				p.Report(g.Grantee, id, "r", "", now)
			}
		case 5:
			_, amount, err := p.Claim(id, now)
			if err == nil {
				if !mayPay(p.Rules(), before, now.Unix()) {
					t.Fatalf("step %d: grant %d paid before it was due", step, id)
				}
				if amount != before.Tranches[len(before.PaidAt)].Amount {
					t.Fatalf("step %d: grant %d paid %d, not its next tranche", step, id, amount)
				}
				balance -= amount
			}
		case 6:
			p.Cancel(id, "stop", now)
		case 7:
			p.Expire(id, now)
		case 9:
			if g, ok := p.Get(id); ok {
				p.Withdraw(g.Grantee, id, now)
			}
		case 8:
			// Refs below the grant's last one are stale and must be refused.
			ref := int64(rnd(step + 1))
			err := p.Hold(id, ref, now)
			if n := len(before.Holds); err == nil && n > 0 && ref <= before.Holds[n-1].Ref {
				t.Fatalf("step %d: grant %d took stale hold ref %d", step, id, ref)
			}
		}
		if step == 350 {
			p.Close("end", seed%2 == 0, now)
		}
		check(t, p, step, balance, final)
	}
}

func check(t *testing.T, p *Program, step int, balance int64, final map[int]string) {
	var committed int64
	open := 0
	pending := map[string]int{}
	for _, g := range p.Page(0, p.Size()) {
		st := g.Status.String()
		if was, ok := final[g.ID]; ok && was != st {
			t.Fatalf("step %d: closed grant %d moved from %s to %s", step, g.ID, was, st)
		}
		switch g.Status {
		case Active:
			committed += g.Total() - g.Paid()
			open++
		case Pending:
			open++
			pending[g.Grantee]++
			if pending[g.Grantee] > 1 {
				t.Fatalf("step %d: %s has two pending requests", step, g.Grantee)
			}
		default:
			final[g.ID] = st
		}
	}
	if committed != p.Committed() {
		t.Fatalf("step %d: committed is %d, active grants still owe %d", step, p.Committed(), committed)
	}
	if balance < committed {
		t.Fatalf("step %d: pool %d below committed %d", step, balance, committed)
	}
	if open != p.ActiveCount()+p.WaitingCount() {
		t.Fatalf("step %d: indexes hold %d grants, %d are open", step, p.ActiveCount()+p.WaitingCount(), open)
	}
	// With the counts equal, right statuses in strictly increasing order
	// means each index holds exactly its grants.
	for _, idx := range []struct {
		list []Grant
		want Status
	}{{p.Active(0, p.ActiveCount()), Active}, {p.Waiting(0, p.WaitingCount()), Pending}} {
		for i, g := range idx.list {
			if g.Status != idx.want || (i > 0 && g.ID <= idx.list[i-1].ID) {
				t.Fatalf("step %d: index for %s is wrong at grant %d", step, idx.want.String(), g.ID)
			}
		}
	}
}

// mayPay works out from g's history, not from payableAt, whether its next
// tranche may be paid at now: a report on that tranche (approval, for an
// advance) a window old, the period since the previous payment, and a hold's
// extra time.
func mayPay(r Rules, g Grant, now int64) bool {
	if g.Status != Active {
		return false
	}
	next := len(g.PaidAt)
	since := g.ApprovedAt
	if next > 0 || !g.Advance {
		since = -1
		for _, rp := range g.Reports {
			if rp.Tranche == next {
				since = rp.At
			}
		}
		if since < 0 {
			return false
		}
	}
	if now < since+secs(r.Window) || next > 0 && now < g.PaidAt[next-1]+secs(g.Period) {
		return false
	}
	for _, h := range g.Holds {
		if h.Tranche == next && now < h.At+secs(r.HoldFor) {
			return false
		}
	}
	return true
}
