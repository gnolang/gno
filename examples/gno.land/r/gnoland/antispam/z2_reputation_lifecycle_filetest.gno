package main

// Demonstration: Reputation Impact on Scoring
// ============================================
// Shows how reputation data affects spam scores. Callers provide chain
// data (AccountAgeDays, Balance, HasUsername) and moderation counters
// (FlaggedCount, BanCount, TotalAccepted) via ReputationData.
//
// When using the shared realm (r/gnoland/antispam), FlaggedCount, BanCount,
// and TotalAccepted are auto-populated from internal state. Callers only
// need to provide the chain data fields.
//
// Scenarios:
//   1. Established user with good history -> no reputation penalties
//   2. Same content from brand new account -> NEW_ACCOUNT + NO_USERNAME + LOW_BALANCE
//   3. Flagged user (flag ratio > 30%) -> BAD_REPUTATION
//   4. Banned user (ban history) -> BANNED_BEFORE
//   5. Combined: new + flagged + banned -> all reputation rules fire

import (
	"gno.land/r/gnoland/antispam"

	engine "gno.land/p/gnoland/antispam"
)

func intToString(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	s := ""
	for i > 0 {
		s = string(rune(48+(i%10))) + s
		i /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

func sliceToString(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	s := "["
	for i, v := range arr {
		if i > 0 {
			s += " "
		}
		s += v
	}
	s += "]"
	return s
}

func main() {
	println("=== REPUTATION IMPACT DEMO ===\n")

	// Same neutral content for all scenarios to isolate reputation effects
	content := "This is a normal message about the latest governance proposal and validator staking options."
	normalRate := engine.RateState{PostCount: 2, WindowSeconds: 3600}

	// ---- Scenario 1: Established user with good history

	println("SCENARIO 1: Established User (Good History)")
	println("---------------------------------------------")

	result1 := antispam.Score(
		address("g1established"),
		content,
		normalRate,
		engine.ReputationData{
			AccountAgeDays: 180,
			Balance:        5000000000, // 5000 GNOT
			HasUsername:    true,
			TotalAccepted:  50,
			FlaggedCount:   0,
			BanCount:       0,
		},
		nil, nil, nil, nil,
		engine.EarlyExitDisabled,
	)

	println("Reputation: 180 days, 5000 GNOT, username, 50 accepted, 0 flags")
	println("Score:      " + intToString(result1.Total))
	println("Triggered:  " + sliceToString(result1.Triggered))
	println("Verdict:    CLEAN - no reputation penalties\n")

	// ---- Scenario 2: Brand new account

	println("SCENARIO 2: Brand New Account")
	println("-------------------------------")

	result2 := antispam.Score(
		address("g1newbie"),
		content,
		normalRate,
		engine.ReputationData{
			AccountAgeDays: 0,
			Balance:        100000000, // 100 GNOT (below 1000 threshold)
			HasUsername:    false,
			TotalAccepted:  0,
		},
		nil, nil, nil, nil,
		engine.EarlyExitDisabled,
	)

	println("Reputation: 0 days, 100 GNOT, no username, 0 accepted")
	println("Score:      " + intToString(result2.Total))
	println("Triggered:  " + sliceToString(result2.Triggered))
	println("NEW_ACCOUNT(2) + NO_USERNAME(1) + LOW_BALANCE(1) = 4 points\n")

	// ---- Scenario 3: Flagged user (bad flag ratio)

	println("SCENARIO 3: Flagged User (Bad Ratio)")
	println("--------------------------------------")

	result3 := antispam.Score(
		address("g1flagged"),
		content,
		normalRate,
		engine.ReputationData{
			AccountAgeDays: 60,
			Balance:        2000000000, // 2000 GNOT
			HasUsername:    true,
			TotalAccepted:  10,
			FlaggedCount:   5, // 50% flagged (threshold: 30%)
		},
		nil, nil, nil, nil,
		engine.EarlyExitDisabled,
	)

	println("Reputation: 60 days, 2000 GNOT, username, 10 accepted, 5 flags (50%)")
	println("Score:      " + intToString(result3.Total))
	println("Triggered:  " + sliceToString(result3.Triggered))
	println("BAD_REPUTATION(3) - flag ratio 50% exceeds 30% threshold\n")

	// ---- Scenario 4: Previously banned user

	println("SCENARIO 4: Previously Banned User")
	println("------------------------------------")

	result4 := antispam.Score(
		address("g1banned"),
		content,
		normalRate,
		engine.ReputationData{
			AccountAgeDays: 90,
			Balance:        3000000000, // 3000 GNOT
			HasUsername:    true,
			TotalAccepted:  20,
			FlaggedCount:   2, // 10% flagged (below threshold)
			BanCount:       2,
		},
		nil, nil, nil, nil,
		engine.EarlyExitDisabled,
	)

	println("Reputation: 90 days, 3000 GNOT, username, 20 accepted, 2 flags, 2 bans")
	println("Score:      " + intToString(result4.Total))
	println("Triggered:  " + sliceToString(result4.Triggered))
	println("BANNED_BEFORE(2) - +1 per past ban, capped at 3\n")

	// ---- Scenario 5: Worst case - new + flagged + banned

	println("SCENARIO 5: Combined Reputation Penalties")
	println("-------------------------------------------")

	result5 := antispam.Score(
		address("g1worstcase"),
		content,
		normalRate,
		engine.ReputationData{
			AccountAgeDays: 0,
			Balance:        50000000, // 50 GNOT
			HasUsername:    false,
			TotalAccepted:  0,
			FlaggedCount:   3, // flags with no accepted = suspicious
			BanCount:       3,
		},
		nil, nil, nil, nil,
		engine.EarlyExitDisabled,
	)

	println("Reputation: 0 days, 50 GNOT, no username, 0 accepted, 3 flags, 3 bans")
	println("Score:      " + intToString(result5.Total))
	println("Triggered:  " + sliceToString(result5.Triggered))
	if result5.Total >= engine.ThresholdReject {
		println("Verdict:    REJECTED (score >= " + intToString(engine.ThresholdReject) + ")")
	}
	println("")

	// ---- Summary

	println("SUMMARY: Same Content, Different Reputation")
	println("---------------------------------------------")
	println("Established:  score=" + intToString(result1.Total) + " (clean)")
	println("New account:  score=" + intToString(result2.Total) + " (caution)")
	println("Flagged:      score=" + intToString(result3.Total) + " (bad ratio)")
	println("Banned:       score=" + intToString(result4.Total) + " (ban history)")
	println("Combined:     score=" + intToString(result5.Total) + " (all penalties)")
}

// Output:
// === REPUTATION IMPACT DEMO ===
//
// SCENARIO 1: Established User (Good History)
// ---------------------------------------------
// Reputation: 180 days, 5000 GNOT, username, 50 accepted, 0 flags
// Score:      0
// Triggered:  []
// Verdict:    CLEAN - no reputation penalties
//
// SCENARIO 2: Brand New Account
// -------------------------------
// Reputation: 0 days, 100 GNOT, no username, 0 accepted
// Score:      4
// Triggered:  [NEW_ACCOUNT NO_USERNAME LOW_BALANCE]
// NEW_ACCOUNT(2) + NO_USERNAME(1) + LOW_BALANCE(1) = 4 points
//
// SCENARIO 3: Flagged User (Bad Ratio)
// --------------------------------------
// Reputation: 60 days, 2000 GNOT, username, 10 accepted, 5 flags (50%)
// Score:      3
// Triggered:  [BAD_REPUTATION]
// BAD_REPUTATION(3) - flag ratio 50% exceeds 30% threshold
//
// SCENARIO 4: Previously Banned User
// ------------------------------------
// Reputation: 90 days, 3000 GNOT, username, 20 accepted, 2 flags, 2 bans
// Score:      2
// Triggered:  [BANNED_BEFORE]
// BANNED_BEFORE(2) - +1 per past ban, capped at 3
//
// SCENARIO 5: Combined Reputation Penalties
// -------------------------------------------
// Reputation: 0 days, 50 GNOT, no username, 0 accepted, 3 flags, 3 bans
// Score:      10
// Triggered:  [NEW_ACCOUNT NO_USERNAME LOW_BALANCE BAD_REPUTATION BANNED_BEFORE]
// Verdict:    REJECTED (score >= 8)
//
// SUMMARY: Same Content, Different Reputation
// ---------------------------------------------
// Established:  score=0 (clean)
// New account:  score=4 (caution)
// Flagged:      score=3 (bad ratio)
// Banned:       score=2 (ban history)
// Combined:     score=10 (all penalties)
