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

func TestCalculatePayoutOverflowingPoolPanics(t *testing.T) {
	const bigPool = int64(3037000500)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on overflowing pool")
		}
	}()
	CalculatePayout(bigPool, bigPool, bigPool)
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

func TestCalculatePayoutWithRakeWinnerTakesAllDoesNotPanic(t *testing.T) {
	payout := CalculatePayoutWithRake(1000, 1000, 1000, 500)
	if payout != 950 {
		t.Fatalf("expected 950, got %d", payout)
	}
}

func TestCalculatePayoutWithRakeWinningPoolAboveNetPoolStillPaysOut(t *testing.T) {
	payout := CalculatePayoutWithRake(951, 1000, 951, 500)
	if payout != 950 {
		t.Fatalf("expected 950, got %d", payout)
	}
}

func TestCalculatePayoutWithRakeNegativeTotalPoolPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on negative totalPool")
		}
	}()
	CalculatePayoutWithRake(100, -1000, 400, 0)
}

func TestCalculateRakeBasic(t *testing.T) {
	rake := CalculateRake(1000, 500)
	if rake != 50 {
		t.Fatalf("expected 50, got %d", rake)
	}
}

func TestCalculateRakeInvalidBpsPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on rakeBps > 10000")
		}
	}()
	CalculateRake(1000, 10001)
}

func TestImpliedProbabilityBpsHalfPool(t *testing.T) {
	prob := ImpliedProbabilityBps(500, 1000)
	if prob != 5000 {
		t.Fatalf("expected 5000, got %d", prob)
	}
}

func TestImpliedProbabilityBpsExactSplitSumsToTotal(t *testing.T) {
	a := ImpliedProbabilityBps(500, 1000)
	b := ImpliedProbabilityBps(300, 1000)
	c := ImpliedProbabilityBps(200, 1000)
	if a+b+c != 10000 {
		t.Fatalf("expected probabilities to sum to 10000, got %d", a+b+c)
	}
}

func TestImpliedProbabilityBpsRoundingLeavesDust(t *testing.T) {
	a := ImpliedProbabilityBps(1000, 3000)
	b := ImpliedProbabilityBps(1000, 3000)
	c := ImpliedProbabilityBps(1000, 3000)
	if a != 3333 || b != 3333 || c != 3333 {
		t.Fatalf("expected 3333 each, got %d, %d, %d", a, b, c)
	}
	if a+b+c != 9999 {
		t.Fatalf("expected rounding dust to leave 9999, got %d", a+b+c)
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

func TestImpliedProbabilityBpsOverflowingPoolPanics(t *testing.T) {
	const bigPool = int64(922337203685478)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on overflowing pool")
		}
	}()
	ImpliedProbabilityBps(bigPool, bigPool)
}
