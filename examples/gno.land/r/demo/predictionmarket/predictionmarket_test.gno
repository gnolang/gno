package predictionmarket

import (
	"strconv"
	"testing"
)

func TestCreateMarket(cur realm, t *testing.T) {
	id := CreateMarket(cur, "Test Market: Team A vs Team B")
	if id != 1 {
		t.Fatalf("expected market id 1, got %d", id)
	}

	raw, exists := markets.Get(strconv.Itoa(id))
	if !exists {
		t.Fatal("market was not stored")
	}
	m := raw.(*Market)
	if m.description != "Test Market: Team A vs Team B" {
		t.Fatalf("unexpected description: %s", m.description)
	}
	if m.resolved {
		t.Fatal("new market should not be resolved")
	}
}

func TestResolveMarket(cur realm, t *testing.T) {
	id := CreateMarket(cur, "Test Market: Resolve Flow")
	ResolveMarket(cur, id, OutcomeHome)

	raw, _ := markets.Get(strconv.Itoa(id))
	m := raw.(*Market)
	if !m.resolved {
		t.Fatal("market should be resolved")
	}
	if m.winner != OutcomeHome {
		t.Fatalf("expected winner %d, got %d", OutcomeHome, m.winner)
	}
}

func TestResolveMarketTwiceFails(cur realm, t *testing.T) {
	id := CreateMarket(cur, "Test Market: Double Resolve")
	ResolveMarket(cur, id, OutcomeAway)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on double resolve")
		}
	}()
	ResolveMarket(cur, id, OutcomeDraw)
}

func TestClaimWinningsBeforeResolveFails(cur realm, t *testing.T) {
	id := CreateMarket(cur, "Test Market: Early Claim")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on claim before resolve")
		}
	}()
	ClaimWinnings(cur, id)
}
