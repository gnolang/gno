package main

// Demonstration: Scoring Content via the Shared Realm
// ====================================================
// Shows how any realm calls Score() to evaluate content.
// No admin access needed - Score() is public.
//
// The realm auto-populates reputation data (FlaggedCount, BanCount,
// TotalAccepted) from internal state. Callers only provide chain
// data (AccountAgeDays, Balance, HasUsername).
//
// IMPORTANT: Score() is READ-ONLY and does NOT update reputation.
// The calling realm must manually record moderation decisions via
// RecordAccepted(), RecordFlag(), and RecordBan() (requires registration).
//
// Scenarios:
//   1. Clean governance article from established user -> score 0
//   2. Spam indicators from new account -> reputation penalties
//   3. Comparison: same content, different reputation -> different scores
//   4. Recording moderation decisions after scoring

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
	println("=== SCORE DEMO: SHARED REALM ===\n")

	// Scenario 1: Established user posting a governance article

	println("SCENARIO 1: Governance Article (Established User)")
	println("--------------------------------------------------")

	// Caller provides chain data only.
	// FlaggedCount, BanCount, TotalAccepted are auto-populated
	// from internal reputation state (zero for unknown address).
	established := engine.ReputationData{
		AccountAgeDays: 180,
		Balance:        5000000000, // 5000 GNOT
		HasUsername:    true,
	}

	result1 := antispam.Score(
		address("g1contributor"),
		"The governance proposal for increasing validator delegation limits has been submitted for community review. Voting starts next week and runs for 7 days. Please review the technical specification before casting your vote on this important change.",
		engine.RateState{PostCount: 2, WindowSeconds: 3600},
		established,
		nil, nil, nil, nil, // use shared state (empty on fresh deploy)
		engine.EarlyExitDisabled,
	)

	println("Author:    g1contributor (180 days, 5000 GNOT, username)")
	println("Content:   Governance proposal (250 chars)")
	println("Score:     " + intToString(result1.Total))
	println("Triggered: " + sliceToString(result1.Triggered))
	println("Verdict:   CLEAN - legitimate content from established user\n")

	// Scenario 2: New account with suspicious indicators

	println("SCENARIO 2: Content from New Account")
	println("--------------------------------------")

	newAccount := engine.ReputationData{
		AccountAgeDays: 0,
		Balance:        100000000, // 100 GNOT (below 1000 threshold)
		HasUsername:    false,
	}

	result2 := antispam.Score(
		address("g1newuser"),
		"THIS IS ALL CAPS AND VERY LOUD AND ANNOYING CONTENT THAT SCREAMS AT EVERYONE!!!",
		engine.RateState{PostCount: 25, WindowSeconds: 3600},
		newAccount,
		nil, nil, nil, nil,
		engine.EarlyExitDisabled,
	)

	println("Author:    g1newuser (0 days, 100 GNOT, no username)")
	println("Content:   All caps shouting (80 chars)")
	println("Score:     " + intToString(result2.Total))
	println("Triggered: " + sliceToString(result2.Triggered))
	if result2.Total >= engine.ThresholdHide {
		println("Verdict:   HIDDEN (score >= " + intToString(engine.ThresholdHide) + ")\n")
	}

	// Scenario 3: Same content, different reputation

	println("SCENARIO 3: Reputation Changes Everything")
	println("-------------------------------------------")

	// Same content but from an established user -> lower score
	result3 := antispam.Score(
		address("g1veteran"),
		"THIS IS ALL CAPS AND VERY LOUD AND ANNOYING CONTENT THAT SCREAMS AT EVERYONE!!!",
		engine.RateState{PostCount: 2, WindowSeconds: 3600},
		established,
		nil, nil, nil, nil,
		engine.EarlyExitDisabled,
	)

	println("Same content, established user (180 days, 5000 GNOT, username)")
	println("Score:     " + intToString(result3.Total))
	println("Triggered: " + sliceToString(result3.Triggered))
	println("Difference: " + intToString(result2.Total-result3.Total) + " points less (no reputation penalties)\n")

	// Scenario 4: Recording moderation decisions
	// Note: This demonstration shows the API, but cannot execute because
	// we're not running as a registered trusted caller realm.

	println("SCENARIO 4: Manual Reputation Recording Workflow")
	println("-------------------------------------------------")
	println("Score() is READ-ONLY - it does NOT automatically update reputation.")
	println("The calling realm must manually record moderation decisions:\n")

	println("// After scoring Scenario 1 (clean content, score 0):")
	println("if result1.Total < engine.ThresholdHide {")
	println("    // Content accepted - record for reputation tracking")
	println("    antispam.RecordAccepted(address(\"g1contributor\"))")
	println("}\n")

	println("// After scoring Scenario 2 (spam, score 10):")
	println("if result2.Total >= engine.ThresholdReject {")
	println("    // Auto-reject high score content")
	println("    // (RecordFlag/RecordBan requires human moderation)")
	println("} else if result2.Total >= engine.ThresholdHide && moderatorFlagged {")
	println("    // Moderator reviewed and confirmed spam")
	println("    antispam.RecordFlag(address(\"g1newuser\"))")
	println("}\n")

	println("if adminBanned {")
	println("    // Admin banned the address (permanent history)")
	println("    antispam.RecordBan(address(\"g1spammer\"))")
	println("}\n")

	println("Rationale: Manual recording gives the realm full control over")
	println("moderation decisions, prevents false positives from automatically")
	println("poisoning reputation data, and separates detection from action.")
}

// Output:
// === SCORE DEMO: SHARED REALM ===
//
// SCENARIO 1: Governance Article (Established User)
// --------------------------------------------------
// Author:    g1contributor (180 days, 5000 GNOT, username)
// Content:   Governance proposal (250 chars)
// Score:     0
// Triggered: []
// Verdict:   CLEAN - legitimate content from established user
//
// SCENARIO 2: Content from New Account
// --------------------------------------
// Author:    g1newuser (0 days, 100 GNOT, no username)
// Content:   All caps shouting (80 chars)
// Score:     10
// Triggered: [RATE_BURST NEW_ACCOUNT NO_USERNAME LOW_BALANCE ALL_CAPS]
// Verdict:   HIDDEN (score >= 5)
//
// SCENARIO 3: Reputation Changes Everything
// -------------------------------------------
// Same content, established user (180 days, 5000 GNOT, username)
// Score:     2
// Triggered: [ALL_CAPS]
// Difference: 8 points less (no reputation penalties)
//
// SCENARIO 4: Manual Reputation Recording Workflow
// -------------------------------------------------
// Score() is READ-ONLY - it does NOT automatically update reputation.
// The calling realm must manually record moderation decisions:
//
// // After scoring Scenario 1 (clean content, score 0):
// if result1.Total < engine.ThresholdHide {
//     // Content accepted - record for reputation tracking
//     antispam.RecordAccepted(address("g1contributor"))
// }
//
// // After scoring Scenario 2 (spam, score 10):
// if result2.Total >= engine.ThresholdReject {
//     // Auto-reject high score content
//     // (RecordFlag/RecordBan requires human moderation)
// } else if result2.Total >= engine.ThresholdHide && moderatorFlagged {
//     // Moderator reviewed and confirmed spam
//     antispam.RecordFlag(address("g1newuser"))
// }
//
// if adminBanned {
//     // Admin banned the address (permanent history)
//     antispam.RecordBan(address("g1spammer"))
// }
//
// Rationale: Manual recording gives the realm full control over
// moderation decisions, prevents false positives from automatically
// poisoning reputation data, and separates detection from action.
