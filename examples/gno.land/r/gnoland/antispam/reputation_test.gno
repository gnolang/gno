package antispam

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"

	engine "gno.land/p/gnoland/antispam"
)

// ---- RecordAccepted tests

func TestRecordAccepted(t *testing.T) {
	resetState(t)
	setAdmin(t)
	// Register test realm as trusted caller
	testRealm := testing.NewCodeRealm("gno.land/r/test/reputation")
	AdminRegisterCaller(cross, testRealm.Address())

	user := address("g1poster")
	testing.SetRealm(testRealm)

	t.Run("success: increments accepted count", func(t *testing.T) {
		RecordAccepted(cross, user)
		RecordAccepted(cross, user)

		rep := GetReputation(user)
		if rep.TotalAccepted != 2 {
			t.Errorf("expected TotalAccepted=2, got %d", rep.TotalAccepted)
		}
	})

	t.Run("success: independent per address", func(t *testing.T) {
		resetState(t)
		setAdmin(t)
		// Register test realm as trusted caller
		testRealmInner := testing.NewCodeRealm("gno.land/r/test/reputation")
		AdminRegisterCaller(cross, testRealmInner.Address())
		testing.SetRealm(testRealmInner)

		RecordAccepted(cross, address("g1a"))
		RecordAccepted(cross, address("g1a"))
		RecordAccepted(cross, address("g1b"))

		repA := GetReputation(address("g1a"))
		repB := GetReputation(address("g1b"))
		if repA.TotalAccepted != 2 {
			t.Errorf("expected g1a TotalAccepted=2, got %d", repA.TotalAccepted)
		}
		if repB.TotalAccepted != 1 {
			t.Errorf("expected g1b TotalAccepted=1, got %d", repB.TotalAccepted)
		}
	})
}

// ---- RecordFlag tests

func TestRecordFlag(t *testing.T) {
	resetState(t)
	setAdmin(t)
	// Register test realm as trusted caller
	testRealm := testing.NewCodeRealm("gno.land/r/test/reputation")
	AdminRegisterCaller(cross, testRealm.Address())

	user := address("g1flagged")
	testing.SetRealm(testRealm)

	t.Run("success: increments flag count", func(t *testing.T) {
		RecordFlag(cross, user)
		RecordFlag(cross, user)
		RecordFlag(cross, user)

		rep := GetReputation(user)
		if rep.FlaggedCount != 3 {
			t.Errorf("expected FlaggedCount=3, got %d", rep.FlaggedCount)
		}
	})
}

// ---- RecordBan tests

func TestRecordBan(t *testing.T) {
	resetState(t)
	setAdmin(t)
	// Register test realm as trusted caller
	testRealm := testing.NewCodeRealm("gno.land/r/test/reputation")
	AdminRegisterCaller(cross, testRealm.Address())

	user := address("g1banned")
	testing.SetRealm(testRealm)

	t.Run("success: increments ban count", func(t *testing.T) {
		RecordBan(cross, user)
		rep := GetReputation(user)
		if rep.BanCount != 1 {
			t.Errorf("expected BanCount=1, got %d", rep.BanCount)
		}
	})

	t.Run("success: multiple bans accumulate", func(t *testing.T) {
		RecordBan(cross, user)
		rep := GetReputation(user)
		if rep.BanCount != 2 {
			t.Errorf("expected BanCount=2, got %d", rep.BanCount)
		}
	})
}

// ---- GetReputation tests

func TestGetReputation(t *testing.T) {
	resetState(t)
	setAdmin(t)
	// Register test realm as trusted caller
	testRealm := testing.NewCodeRealm("gno.land/r/test/reputation")
	AdminRegisterCaller(cross, testRealm.Address())
	testing.SetRealm(testRealm)

	t.Run("success: unknown address returns zero", func(t *testing.T) {
		rep := GetReputation(address("g1unknown"))
		if rep.TotalAccepted != 0 || rep.FlaggedCount != 0 || rep.BanCount != 0 {
			t.Errorf("expected all zeros for unknown address, got accepted=%d flags=%d bans=%d",
				rep.TotalAccepted, rep.FlaggedCount, rep.BanCount)
		}
	})

	t.Run("success: returns accumulated data", func(t *testing.T) {
		user := address("g1full")
		RecordAccepted(cross, user)
		RecordAccepted(cross, user)
		RecordAccepted(cross, user)
		RecordFlag(cross, user)
		RecordBan(cross, user)

		rep := GetReputation(user)
		if rep.TotalAccepted != 3 {
			t.Errorf("expected TotalAccepted=3, got %d", rep.TotalAccepted)
		}
		if rep.FlaggedCount != 1 {
			t.Errorf("expected FlaggedCount=1, got %d", rep.FlaggedCount)
		}
		if rep.BanCount != 1 {
			t.Errorf("expected BanCount=1, got %d", rep.BanCount)
		}
	})
}

// ---- Trusted caller access control

func TestTrustedCallerAccess(t *testing.T) {
	resetState(t)

	// Register a realm as trusted caller
	callerRealm := testing.NewCodeRealm("gno.land/r/test/boards")
	AdminRegisterCaller(cross, callerRealm.Address())

	// Switch to calling from the registered realm
	testing.SetRealm(callerRealm)

	// Registered caller can record
	RecordAccepted(cross, address("g1user"))
	RecordFlag(cross, address("g1user"))
	RecordBan(cross, address("g1user"))

	rep := GetReputation(address("g1user"))
	if rep.TotalAccepted != 1 || rep.FlaggedCount != 1 || rep.BanCount != 1 {
		t.Errorf("expected 1/1/1, got accepted=%d flags=%d bans=%d",
			rep.TotalAccepted, rep.FlaggedCount, rep.BanCount)
	}
}

func TestUntrustedRealmRejected(t *testing.T) {
	resetState(t)

	// Switch to an unregistered realm
	untrusted := testing.NewCodeRealm("gno.land/r/evil/realm")
	testing.SetRealm(untrusted)

	uassert.AbortsWithMessage(t, errUnauthorized.Error(), func() {
		RecordAccepted(cross, address("g1user"))
	})
}

func TestUntrustedUserRejected(t *testing.T) {
	resetState(t)

	// Switch to a non-admin user
	other := testutils.TestAddress("other")
	testing.SetOriginCaller(other)
	testing.SetRealm(testing.NewUserRealm(other))

	uassert.AbortsWithMessage(t, errUnauthorized.Error(), func() {
		RecordAccepted(cross, address("g1user"))
	})
}

// ---- Score integration with reputation

func TestScoreUsesInternalReputation(t *testing.T) {
	resetState(t)
	setAdmin(t)
	// Register test realm as trusted caller
	testRealm := testing.NewCodeRealm("gno.land/r/test/reputation")
	AdminRegisterCaller(cross, testRealm.Address())
	testing.SetRealm(testRealm)

	user := address("g1spammy")

	t.Run("success: bad flag ratio triggers BAD_REPUTATION", func(t *testing.T) {
		// Create a user with 5 accepted, 3 flagged -> 60% flag ratio > 30%
		for i := 0; i < 5; i++ {
			RecordAccepted(cross, user)
		}
		for i := 0; i < 3; i++ {
			RecordFlag(cross, user)
		}

		// Caller provides chain data only, internal rep is auto-populated
		result := Score(user, "Normal message from a flagged user with enough content.",
			engine.RateState{},
			engine.ReputationData{
				AccountAgeDays: 30,
				Balance:        5000000000,
				HasUsername:    true,
			},
			nil, nil, nil, nil, engine.EarlyExitDisabled,
		)

		hasBadRep := false
		for _, r := range result.Triggered {
			if r == "BAD_REPUTATION" {
				hasBadRep = true
			}
		}
		if !hasBadRep {
			t.Errorf("expected BAD_REPUTATION for 60%% flag ratio, got triggered: %v (score: %d)",
				result.Triggered, result.Total)
		}
	})

	t.Run("success: ban history triggers BANNED_BEFORE", func(t *testing.T) {
		resetState(t)
		setAdmin(t)
		// Register test realm as trusted caller
		testRealmInner := testing.NewCodeRealm("gno.land/r/test/reputation")
		AdminRegisterCaller(cross, testRealmInner.Address())
		testing.SetRealm(testRealmInner)

		banned := address("g1exbanned")
		RecordBan(cross, banned)

		result := Score(banned, "Content again after ban with enough for analysis here.",
			engine.RateState{},
			engine.ReputationData{
				AccountAgeDays: 30,
				Balance:        5000000000,
				HasUsername:    true,
			},
			nil, nil, nil, nil, engine.EarlyExitDisabled,
		)

		hasBanned := false
		for _, r := range result.Triggered {
			if r == "BANNED_BEFORE" {
				hasBanned = true
			}
		}
		if !hasBanned {
			t.Errorf("expected BANNED_BEFORE for ban count=1, got triggered: %v (score: %d)",
				result.Triggered, result.Total)
		}
	})

	t.Run("success: clean user has no reputation penalties", func(t *testing.T) {
		resetState(t)
		setAdmin(t)
		// Register test realm as trusted caller
		testRealmInner := testing.NewCodeRealm("gno.land/r/test/reputation")
		AdminRegisterCaller(cross, testRealmInner.Address())
		testing.SetRealm(testRealmInner)

		clean := address("g1clean")
		for i := 0; i < 10; i++ {
			RecordAccepted(cross, clean)
		}

		result := Score(clean, "This is a perfectly normal discussion about governance proposals.",
			engine.RateState{},
			engine.ReputationData{
				AccountAgeDays: 60,
				Balance:        5000000000,
				HasUsername:    true,
			},
			nil, nil, nil, nil, engine.EarlyExitDisabled,
		)

		if result.Total != 0 {
			t.Errorf("expected score=0 for clean user, got %d (triggered: %v)",
				result.Total, result.Triggered)
		}
	})
}

// ---- AdminResetReputation

func TestAdminResetReputation(t *testing.T) {
	resetState(t)
	setAdmin(t)
	// Register test realm as trusted caller
	testRealm := testing.NewCodeRealm("gno.land/r/test/reputation")
	AdminRegisterCaller(cross, testRealm.Address())
	testing.SetRealm(testRealm)

	user := address("g1resetme")
	RecordAccepted(cross, user)
	RecordAccepted(cross, user)
	RecordFlag(cross, user)
	RecordBan(cross, user)

	AdminResetReputation(cross, user)

	rep := GetReputation(user)
	if rep.TotalAccepted != 0 || rep.FlaggedCount != 0 || rep.BanCount != 0 {
		t.Errorf("expected all zeros after reset, got accepted=%d flags=%d bans=%d",
			rep.TotalAccepted, rep.FlaggedCount, rep.BanCount)
	}
}

func TestAdminResetReputationNonAdmin(t *testing.T) {
	resetState(t)
	setAdmin(t)

	other := testutils.TestAddress("other")
	testing.SetOriginCaller(other)

	uassert.AbortsWithMessage(t, errUnauthorized.Error(), func() {
		AdminResetReputation(cross, address("g1someone"))
	})
}

// ---- ReputationCount

func TestReputationCount(t *testing.T) {
	resetState(t)
	setAdmin(t)
	// Register test realm as trusted caller
	testRealm := testing.NewCodeRealm("gno.land/r/test/reputation")
	AdminRegisterCaller(cross, testRealm.Address())
	testing.SetRealm(testRealm)

	if ReputationCount() != 0 {
		t.Errorf("expected 0 tracked addresses, got %d", ReputationCount())
	}

	RecordAccepted(cross, address("g1a"))
	RecordAccepted(cross, address("g1b"))
	RecordFlag(cross, address("g1c"))

	if ReputationCount() != 3 {
		t.Errorf("expected 3 tracked addresses, got %d", ReputationCount())
	}
}
