package parimutuel

import "testing"

func TestCalculatePayoutSingleWinnerTakesAll(t *testing.T) {
	payout := CalculatePayout(1000, 1000, 1000)
	if payout != 1000 {
		t.Fatalf("expected 1000, got %d", payout)
	}
}

func TestCalculatePayoutSplitsProportionally(t *testing.T) {
	payout := CalculatePayout(300, 1000, 400)
	if payout != 750 {
		t.Fatalf("expected 750, got %d", payout)
	}
}

func TestCalculatePayoutZeroWinningPoolPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero winningPool")
		}
	}()
	CalculatePayout(100, 1000, 0)
}

func TestCalculatePayoutWinningPoolExceedsTotalPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when winningPool > totalPool")
		}
	}()
	CalculatePayout(100, 500, 600)
}

func TestCalculatePayoutUserBetExceedsWinningPoolPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when userBet > winningPool")
		}
	}()
	CalculatePayout(600, 1000, 500)
}

func TestCalculatePayoutWithRakeZeroRakeMatchesNoRake(t *testing.T) {
	withoutRake := CalculatePayout(300, 1000, 400)
	withZeroRake := CalculatePayoutWithRake(300, 1000, 400, 0)
	if withoutRake != withZeroRake {
		t.Fatalf("expected %d, got %d", withoutRake, withZeroRake)
	}
}

func TestCalculatePayoutWithRakeReducesPayout(t *testing.T) {
	payout := CalculatePayoutWithRake(400, 1000, 400, 500)
	if payout != 950 {
		t.Fatalf("expected 950, got %d", payout)
	}
}

func TestCalculatePayoutWithRakeFullRakeYieldsZero(t *testing.T) {
	payout := CalculatePayoutWithRake(400, 1000, 400, 10000)
	if payout != 0 {
		t.Fatalf("expected 0, got %d", payout)
	}
}

func TestCalculatePayoutWithRakeInvalidBpsPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on rakeBps > 10000")
		}
	}()
	CalculatePayoutWithRake(100, 1000, 400, 10001)
}

func TestImpliedProbabilityBpsHalfPool(t *testing.T) {
	prob := ImpliedProbabilityBps(500, 1000)
	if prob != 5000 {
		t.Fatalf("expected 5000, got %d", prob)
	}
}

func TestImpliedProbabilityBpsSumsToTotal(t *testing.T) {
	a := ImpliedProbabilityBps(500, 1000)
	b := ImpliedProbabilityBps(300, 1000)
	c := ImpliedProbabilityBps(200, 1000)
	if a+b+c != 10000 {
		t.Fatalf("expected probabilities to sum to 10000, got %d", a+b+c)
	}
}

func TestImpliedProbabilityBpsZeroTotalPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero totalPool")
		}
	}()
	ImpliedProbabilityBps(0, 0)
}

func TestImpliedProbabilityBpsOutcomeExceedsTotalPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when outcomePool > totalPool")
		}
	}()
	ImpliedProbabilityBps(1200, 1000)
}
