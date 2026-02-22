package antispam

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"

	engine "gno.land/p/gnoland/antispam"
)

// ---- RecordPost tests

func TestRecordPost(t *testing.T) {
	resetState(t)
	user := address("g1poster")

	t.Run("success: increments post count", func(t *testing.T) {
		RecordPost(cross, user)
		RecordPost(cross, user)

		rep := GetReputation(user)
		if rep.TotalPosts != 2 {
			t.Errorf("expected TotalPosts=2, got %d", rep.TotalPosts)
		}
	})

	t.Run("success: independent per address", func(t *testing.T) {
		resetState(t)
		RecordPost(cross, address("g1a"))
		RecordPost(cross, address("g1a"))
		RecordPost(cross, address("g1b"))

		repA := GetReputation(address("g1a"))
		repB := GetReputation(address("g1b"))
		if repA.TotalPosts != 2 {
			t.Errorf("expected g1a TotalPosts=2, got %d", repA.TotalPosts)
		}
		if repB.TotalPosts != 1 {
			t.Errorf("expected g1b TotalPosts=1, got %d", repB.TotalPosts)
		}
	})
}

// ---- RecordFlag tests

func TestRecordFlag(t *testing.T) {
	resetState(t)
	user := address("g1flagged")

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

// ---- RecordBan / RecordUnban tests

func TestRecordBan(t *testing.T) {
	resetState(t)
	user := address("g1banned")

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

func TestRecordUnban(t *testing.T) {
	resetState(t)
	user := address("g1unbanned")

	t.Run("success: does not change ban count", func(t *testing.T) {
		RecordBan(cross, user)
		RecordBan(cross, user)
		RecordUnban(cross, user)

		// Unban removes from active ban but ban history persists
		rep := GetReputation(user)
		if rep.BanCount != 2 {
			t.Errorf("expected BanCount=2 (history persists), got %d", rep.BanCount)
		}
	})
}

// ---- GetReputation tests

func TestGetReputation(t *testing.T) {
	resetState(t)

	t.Run("success: unknown address returns zero", func(t *testing.T) {
		rep := GetReputation(address("g1unknown"))
		if rep.TotalPosts != 0 || rep.FlaggedCount != 0 || rep.BanCount != 0 {
			t.Errorf("expected all zeros for unknown address, got posts=%d flags=%d bans=%d",
				rep.TotalPosts, rep.FlaggedCount, rep.BanCount)
		}
	})

	t.Run("success: returns accumulated data", func(t *testing.T) {
		user := address("g1full")
		RecordPost(cross, user)
		RecordPost(cross, user)
		RecordPost(cross, user)
		RecordFlag(cross, user)
		RecordBan(cross, user)

		rep := GetReputation(user)
		if rep.TotalPosts != 3 {
			t.Errorf("expected TotalPosts=3, got %d", rep.TotalPosts)
		}
		if rep.FlaggedCount != 1 {
			t.Errorf("expected FlaggedCount=1, got %d", rep.FlaggedCount)
		}
		if rep.BanCount != 1 {
			t.Errorf("expected BanCount=1, got %d", rep.BanCount)
		}
	})
}

// ---- Score integration with reputation

func TestScoreUsesInternalReputation(t *testing.T) {
	resetState(t)
	user := address("g1spammy")

	t.Run("success: bad flag ratio triggers BAD_REPUTATION", func(t *testing.T) {
		// Create a user with 5 posts, 3 flagged -> 60% flag ratio > 30%
		for i := 0; i < 5; i++ {
			RecordPost(cross, user)
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
		banned := address("g1exbanned")
		RecordBan(cross, banned)

		result := Score(banned, "Posting again after ban with enough content for analysis.",
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
		clean := address("g1clean")
		for i := 0; i < 10; i++ {
			RecordPost(cross, clean)
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

	user := address("g1resetme")
	RecordPost(cross, user)
	RecordPost(cross, user)
	RecordFlag(cross, user)
	RecordBan(cross, user)

	AdminResetReputation(cross, user)

	rep := GetReputation(user)
	if rep.TotalPosts != 0 || rep.FlaggedCount != 0 || rep.BanCount != 0 {
		t.Errorf("expected all zeros after reset, got posts=%d flags=%d bans=%d",
			rep.TotalPosts, rep.FlaggedCount, rep.BanCount)
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

	if ReputationCount() != 0 {
		t.Errorf("expected 0 tracked addresses, got %d", ReputationCount())
	}

	RecordPost(cross, address("g1a"))
	RecordPost(cross, address("g1b"))
	RecordFlag(cross, address("g1c"))

	if ReputationCount() != 3 {
		t.Errorf("expected 3 tracked addresses, got %d", ReputationCount())
	}
}
