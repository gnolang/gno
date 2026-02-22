package main

// Demonstration: EarlyExitAt Gas Optimization
// =============================================
// Rules are ordered by ascending cost with earlyExit checks between
// each group. The same spam message is scored three ways:
//
//   1. Full scoring (EarlyExitDisabled) - all rules evaluated
//   2. Early exit at ThresholdReject (8) - skip costly rules when score >= 8
//   3. Early exit at ThresholdHide (5) - skip costly rules when score >= 5
//
// When cheap rules (rate, reputation) already produce a high score,
// costlier rules (content scan, regex, Bayes, keywords, fingerprints)
// are skipped - saving gas without changing the outcome.
//
// Scenario A: Obvious spammer (cheap rules score >= 8) -> expensive rules skipped
// Scenario B: Subtle spam (cheap rules score < 5) -> expensive rules needed
// Scenario C: Lightweight mode (nil state) -> expensive rules skipped entirely

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

	// Shared setup

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
	// Cheap rules alone score >= 8 (rate + reputation)
	// -> expensive rules skipped with EarlyExitAt=ThresholdReject
	// ========================================

	println("SCENARIO A: Obvious Spammer (cheap rules sufficient)")
	println("=================================================\n")

	badRep := antispam.ReputationData{
		AccountAgeDays: 0,
		Balance:        100000000,
		FlaggedCount:   10,
		TotalAccepted:  15,
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

	// Early exit at ThresholdReject: skip expensive rules
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
	println("  All rules evaluated\n")

	println("Early exit at ThresholdReject (" + intToString(antispam.ThresholdReject) + "):")
	println("  Score: " + intToString(earlyResult.Total) + " | Rules: " + sliceToString(earlyResult.Triggered))
	println("  Cheap rules already >= " + intToString(antispam.ThresholdReject) + " -> expensive rules SKIPPED")
	println("  Outcome identical: REJECTED in both cases, but fewer rules evaluated\n")

	// ========================================
	// SCENARIO B: Subtle Spam
	// Cheap rules score < 5 (good reputation, normal rate)
	// -> expensive rules needed to detect it
	// ========================================

	println("SCENARIO B: Subtle Spam (expensive rules needed)")
	println("=========================================\n")

	goodRep := antispam.ReputationData{
		AccountAgeDays: 60,
		Balance:        5000000000,
		FlaggedCount:   0,
		TotalAccepted:  50,
		HasUsername:    true,
		BanCount:       0,
	}
	normalRate := antispam.RateState{PostCount: 3, WindowSeconds: 3600}
	subtleSpam := "Hey check out this free airdrop bonus claim your prize now"

	// Full scoring: expensive rules catch it
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

	// Early exit at ThresholdReject: score < 8, so expensive rules run too
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
	println("  Expensive rules detect spam vocabulary\n")

	println("Early exit at ThresholdReject (" + intToString(antispam.ThresholdReject) + "):")
	println("  Score: " + intToString(subtleEarly.Total) + " | Rules: " + sliceToString(subtleEarly.Triggered))
	println("  Cheap rules score < " + intToString(antispam.ThresholdReject) + " -> expensive rules RUN (needed to catch subtle spam)\n")

	// ========================================
	// SCENARIO C: Lightweight Mode (nil state)
	// No Corpus/Fps/Bl/Dict -> expensive rules skipped entirely
	// ========================================

	println("SCENARIO C: Lightweight Mode (nil state)")
	println("=========================================\n")

	lightResult := antispam.Score(antispam.ScoreInput{
		Author:  "g1spammer",
		Content: spamMsg,
		Rate:    burstRate,
		Rep:     badRep,
		// Corpus, Fps, Bl, Dict all nil -> expensive rules skipped
	})

	println("Message: \"" + spamMsg + "\"")
	println("Author:  same obvious spammer as Scenario A\n")

	println("Lightweight scoring (all state nil):")
	println("  Score: " + intToString(lightResult.Total) + " | Rules: " + sliceToString(lightResult.Triggered))
	println("  Only cheap rules fire (rate + reputation + content heuristics)")
	println("  No regex, no Bayes, no keywords, no fingerprints")
	println("  Cheapest possible scoring mode\n")

	// ========================================
	// SUMMARY
	// ========================================

	println("SUMMARY")
	println("=======\n")

	println("Mode                          Score  Rules Evaluated")
	println("----                           -   -----------")
	println("A. Full (EarlyExitDisabled)    " + intToString(fullResult.Total) + "  All rules evaluated")
	println("A. EarlyExit at Reject          " + intToString(earlyResult.Total) + "  Cheap rules only (expensive skipped)")
	println("B. Full (subtle spam)           " + intToString(subtleFull.Total) + "  All rules (expensive needed)")
	println("B. EarlyExit at Reject          " + intToString(subtleEarly.Total) + "  All rules (score < threshold)")
	println("C. Lightweight (nil state)      " + intToString(lightResult.Total) + "  Cheap rules only (no state loaded)")
	println("")
	println("EarlyExitAt saves gas on obvious spam without affecting detection of subtle spam.")
	println("Lightweight mode (nil state) is the cheapest option for high-volume realms.")

	println("\n=== END ===")
}

// Output:
// === EARLY EXIT GAS OPTIMIZATION ===
//
// SCENARIO A: Obvious Spammer (cheap rules sufficient)
// =================================================
//
// Message: "FREE MONEY BONUS PRIZE AIRDROP CLICK NOW!!!"
// Author:  new account, no username, low balance, 1 ban, 25 posts/hour
//
// Full scoring (EarlyExitDisabled):
//   Score: 18 | Rules: [RATE_BURST NO_USERNAME LOW_BALANCE BAD_REPUTATION BANNED_BEFORE ALL_CAPS BAYES_SPAM KEYWORD_SPAM]
//   All rules evaluated
//
// Early exit at ThresholdReject (8):
//   Score: 10 | Rules: [RATE_BURST NO_USERNAME LOW_BALANCE BAD_REPUTATION BANNED_BEFORE]
//   Cheap rules already >= 8 -> expensive rules SKIPPED
//   Outcome identical: REJECTED in both cases, but fewer rules evaluated
//
// SCENARIO B: Subtle Spam (expensive rules needed)
// =========================================
//
// Message: "Hey check out this free airdrop bonus claim your prize now"
// Author:  established account, good reputation, normal rate
//
// Full scoring (EarlyExitDisabled):
//   Score: 6 | Rules: [BAYES_SPAM KEYWORD_SPAM]
//   Expensive rules detect spam vocabulary
//
// Early exit at ThresholdReject (8):
//   Score: 6 | Rules: [BAYES_SPAM KEYWORD_SPAM]
//   Cheap rules score < 8 -> expensive rules RUN (needed to catch subtle spam)
//
// SCENARIO C: Lightweight Mode (nil state)
// =========================================
//
// Message: "FREE MONEY BONUS PRIZE AIRDROP CLICK NOW!!!"
// Author:  same obvious spammer as Scenario A
//
// Lightweight scoring (all state nil):
//   Score: 12 | Rules: [RATE_BURST NO_USERNAME LOW_BALANCE BAD_REPUTATION BANNED_BEFORE ALL_CAPS]
//   Only cheap rules fire (rate + reputation + content heuristics)
//   No regex, no Bayes, no keywords, no fingerprints
//   Cheapest possible scoring mode
//
// SUMMARY
// =======
//
// Mode                          Score  Rules Evaluated
//                           -   -----------
// A. Full (EarlyExitDisabled)    18  All rules evaluated
// A. EarlyExit at Reject          10  Cheap rules only (expensive skipped)
// B. Full (subtle spam)           6  All rules (expensive needed)
// B. EarlyExit at Reject          6  All rules (score < threshold)
// C. Lightweight (nil state)      12  Cheap rules only (no state loaded)
//
// EarlyExitAt saves gas on obvious spam without affecting detection of subtle spam.
// Lightweight mode (nil state) is the cheapest option for high-volume realms.
//
// === END ===
