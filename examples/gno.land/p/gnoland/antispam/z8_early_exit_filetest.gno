package main

// Demonstration: EarlyExitAt Gas Optimization
// =============================================
// This test shows how EarlyExitAt controls which scoring phases run.
// The same spam message is scored three ways:
//
//   1. Full scoring (EarlyExitDisabled) - all rules evaluated
//   2. Early exit at ThresholdReject (8) - skip Phase 2 when Phase 1 >= 8
//   3. Early exit at ThresholdHide (5) - skip Phase 2 when Phase 1 >= 5
//
// When cheap Phase 1 rules (content heuristics, rate, reputation) already
// produce a high score, expensive Phase 2 rules (regex, Bayes, keywords,
// fingerprints) are skipped - saving gas without changing the outcome.
//
// Scenario A: Obvious spammer (Phase 1 score >= 8) -> Phase 2 skipped
// Scenario B: Subtle spam (Phase 1 score < 5) -> Phase 2 needed to catch it
// Scenario C: Lightweight mode (nil state) -> Phase 2 skipped entirely

import (
	"gno.land/p/gnoland/antispam"
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
	println("=== EARLY EXIT GAS OPTIMIZATION ===\n")

	// ---- Shared setup

	corpus := antispam.NewCorpus()
	corpus.Train("free airdrop bonus click claim tokens prize", true)
	corpus.Train("free bonus money prize casino jackpot gambling", true)
	corpus.Train("free claim airdrop offer exclusive deal limited", true)
	corpus.Train("governance proposal voting community discussion", false)
	corpus.Train("validator staking delegation reward system", false)
	corpus.Train("development update release version changelog", false)

	dict := antispam.NewKeywordDict()
	dict.BulkAdd("free:3\nairdrop:3\nbonus:3\nclaim:2\nprize:3\ncasino:3\njackpot:3")

	fps := antispam.NewFingerprintStore()
	bl := antispam.NewBlocklist()
	bl.AddPattern(`(?i)send\s+\d+.*get\s+\d+`)

	// ========================================
	// SCENARIO A: Obvious Spammer
	// Phase 1 alone scores >= 8 (new account + caps + rate + reputation)
	// -> Phase 2 skipped with EarlyExitAt=ThresholdReject
	// ========================================

	println("SCENARIO A: Obvious Spammer (Phase 1 sufficient)")
	println("=================================================\n")

	badRep := antispam.ReputationData{
		AccountAgeDays: 0,
		Balance:        100000000,
		FlaggedCount:   10,
		TotalPosts:     15,
		HasUsername:    false,
		BanCount:       1,
	}
	burstRate := antispam.RateState{PostCount: 25, WindowSeconds: 3600}
	spamMsg := "FREE MONEY BONUS PRIZE AIRDROP CLICK NOW!!!"

	// Full scoring: all rules
	fullResult := antispam.Score(antispam.ScoreInput{
		Author:      "g1spammer",
		Content:     spamMsg,
		Rate:        burstRate,
		Rep:         badRep,
		Corpus:      corpus,
		Fps:         fps,
		Bl:          bl,
		Dict:        dict,
		EarlyExitAt: antispam.EarlyExitDisabled,
	})

	// Early exit at ThresholdReject: skip Phase 2
	earlyResult := antispam.Score(antispam.ScoreInput{
		Author:      "g1spammer",
		Content:     spamMsg,
		Rate:        burstRate,
		Rep:         badRep,
		Corpus:      corpus,
		Fps:         fps,
		Bl:          bl,
		Dict:        dict,
		EarlyExitAt: antispam.ThresholdReject,
	})

	println("Message: \"" + spamMsg + "\"")
	println("Author:  new account, no username, low balance, 1 ban, 25 posts/hour\n")

	println("Full scoring (EarlyExitDisabled):")
	println("  Score: " + intToString(fullResult.Total) + " | Rules: " + sliceToString(fullResult.Triggered))
	println("  Phase 1 + Phase 2 both evaluated\n")

	println("Early exit at ThresholdReject (" + intToString(antispam.ThresholdReject) + "):")
	println("  Score: " + intToString(earlyResult.Total) + " | Rules: " + sliceToString(earlyResult.Triggered))
	println("  Phase 1 already >= " + intToString(antispam.ThresholdReject) + " -> Phase 2 SKIPPED")
	println("  Outcome identical: REJECTED in both cases, but fewer rules evaluated\n")

	// ========================================
	// SCENARIO B: Subtle Spam
	// Phase 1 score < 5 (good reputation, normal rate)
	// -> Phase 2 needed to detect it
	// ========================================

	println("SCENARIO B: Subtle Spam (Phase 2 needed)")
	println("=========================================\n")

	goodRep := antispam.ReputationData{
		AccountAgeDays: 60,
		Balance:        5000000000,
		FlaggedCount:   0,
		TotalPosts:     50,
		HasUsername:    true,
		BanCount:       0,
	}
	normalRate := antispam.RateState{PostCount: 3, WindowSeconds: 3600}
	subtleSpam := "Hey check out this free airdrop bonus claim your prize now"

	// Full scoring: Phase 2 catches it
	subtleFull := antispam.Score(antispam.ScoreInput{
		Author:      "g1sneaky",
		Content:     subtleSpam,
		Rate:        normalRate,
		Rep:         goodRep,
		Corpus:      corpus,
		Fps:         fps,
		Bl:          bl,
		Dict:        dict,
		EarlyExitAt: antispam.EarlyExitDisabled,
	})

	// Early exit at ThresholdReject: Phase 1 < 8, so Phase 2 runs too
	subtleEarly := antispam.Score(antispam.ScoreInput{
		Author:      "g1sneaky",
		Content:     subtleSpam,
		Rate:        normalRate,
		Rep:         goodRep,
		Corpus:      corpus,
		Fps:         fps,
		Bl:          bl,
		Dict:        dict,
		EarlyExitAt: antispam.ThresholdReject,
	})

	println("Message: \"" + subtleSpam + "\"")
	println("Author:  established account, good reputation, normal rate\n")

	println("Full scoring (EarlyExitDisabled):")
	println("  Score: " + intToString(subtleFull.Total) + " | Rules: " + sliceToString(subtleFull.Triggered))
	println("  Phase 2 detects spam vocabulary\n")

	println("Early exit at ThresholdReject (" + intToString(antispam.ThresholdReject) + "):")
	println("  Score: " + intToString(subtleEarly.Total) + " | Rules: " + sliceToString(subtleEarly.Triggered))
	println("  Phase 1 score < " + intToString(antispam.ThresholdReject) + " -> Phase 2 RUNS (needed to catch subtle spam)\n")

	// ========================================
	// SCENARIO C: Lightweight Mode (nil state)
	// No Corpus/Fps/Bl/Dict -> Phase 2 skipped entirely
	// ========================================

	println("SCENARIO C: Lightweight Mode (nil state)")
	println("=========================================\n")

	lightResult := antispam.Score(antispam.ScoreInput{
		Author:  "g1spammer",
		Content: spamMsg,
		Rate:    burstRate,
		Rep:     badRep,
		// Corpus, Fps, Bl, Dict all nil -> Phase 2 impossible
	})

	println("Message: \"" + spamMsg + "\"")
	println("Author:  same obvious spammer as Scenario A\n")

	println("Lightweight scoring (all state nil):")
	println("  Score: " + intToString(lightResult.Total) + " | Rules: " + sliceToString(lightResult.Triggered))
	println("  Only Phase 1 rules fire (content heuristics + rate + reputation)")
	println("  No regex, no Bayes, no keywords, no fingerprints")
	println("  Cheapest possible scoring mode\n")

	// ========================================
	// SUMMARY
	// ========================================

	println("SUMMARY")
	println("=======\n")

	println("Mode                          Score  Rules Evaluated")
	println("----                          -----  ---------------")
	println("A. Full (EarlyExitDisabled)    " + intToString(fullResult.Total) + "  Phase 1 + Phase 2 (all rules)")
	println("A. EarlyExit at Reject          " + intToString(earlyResult.Total) + "  Phase 1 only (Phase 2 skipped)")
	println("B. Full (subtle spam)           " + intToString(subtleFull.Total) + "  Phase 1 + Phase 2 (needed)")
	println("B. EarlyExit at Reject          " + intToString(subtleEarly.Total) + "  Phase 1 + Phase 2 (score < threshold)")
	println("C. Lightweight (nil state)      " + intToString(lightResult.Total) + "  Phase 1 only (no state loaded)")
	println("")
	println("EarlyExitAt saves gas on obvious spam without affecting detection of subtle spam.")
	println("Lightweight mode (nil state) is the cheapest option for high-volume realms.")

	println("\n=== END ===")
}

// Output:
// === EARLY EXIT GAS OPTIMIZATION ===
//
// SCENARIO A: Obvious Spammer (Phase 1 sufficient)
// =================================================
//
// Message: "FREE MONEY BONUS PRIZE AIRDROP CLICK NOW!!!"
// Author:  new account, no username, low balance, 1 ban, 25 posts/hour
//
// Full scoring (EarlyExitDisabled):
//   Score: 18 | Rules: [ALL_CAPS RATE_BURST NO_USERNAME LOW_BALANCE BAD_REPUTATION BANNED_BEFORE BAYES_SPAM KEYWORD_SPAM]
//   Phase 1 + Phase 2 both evaluated
//
// Early exit at ThresholdReject (8):
//   Score: 12 | Rules: [ALL_CAPS RATE_BURST NO_USERNAME LOW_BALANCE BAD_REPUTATION BANNED_BEFORE]
//   Phase 1 already >= 8 -> Phase 2 SKIPPED
//   Outcome identical: REJECTED in both cases, but fewer rules evaluated
//
// SCENARIO B: Subtle Spam (Phase 2 needed)
// =========================================
//
// Message: "Hey check out this free airdrop bonus claim your prize now"
// Author:  established account, good reputation, normal rate
//
// Full scoring (EarlyExitDisabled):
//   Score: 6 | Rules: [BAYES_SPAM KEYWORD_SPAM]
//   Phase 2 detects spam vocabulary
//
// Early exit at ThresholdReject (8):
//   Score: 6 | Rules: [BAYES_SPAM KEYWORD_SPAM]
//   Phase 1 score < 8 -> Phase 2 RUNS (needed to catch subtle spam)
//
// SCENARIO C: Lightweight Mode (nil state)
// =========================================
//
// Message: "FREE MONEY BONUS PRIZE AIRDROP CLICK NOW!!!"
// Author:  same obvious spammer as Scenario A
//
// Lightweight scoring (all state nil):
//   Score: 12 | Rules: [ALL_CAPS RATE_BURST NO_USERNAME LOW_BALANCE BAD_REPUTATION BANNED_BEFORE]
//   Only Phase 1 rules fire (content heuristics + rate + reputation)
//   No regex, no Bayes, no keywords, no fingerprints
//   Cheapest possible scoring mode
//
// SUMMARY
// =======
//
// Mode                          Score  Rules Evaluated
// ----                          -----  ---------------
// A. Full (EarlyExitDisabled)    18  Phase 1 + Phase 2 (all rules)
// A. EarlyExit at Reject          12  Phase 1 only (Phase 2 skipped)
// B. Full (subtle spam)           6  Phase 1 + Phase 2 (needed)
// B. EarlyExit at Reject          6  Phase 1 + Phase 2 (score < threshold)
// C. Lightweight (nil state)      12  Phase 1 only (no state loaded)
//
// EarlyExitAt saves gas on obvious spam without affecting detection of subtle spam.
// Lightweight mode (nil state) is the cheapest option for high-volume realms.
//
// === END ===
