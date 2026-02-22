package main

// Demonstration: Gas Benchmark for Full Pipeline
// ===============================================
// Exercises the complete antispam scoring pipeline with realistic data
// volumes for gas profiling. All state containers are populated to
// typical production sizes, then various content types are scored.
//
// State volumes:
//   Corpus: ~100 unique tokens from balanced training
//   FingerprintStore: 50 entries
//   Blocklist: 5 patterns, 10 blocked addresses
//   KeywordDict: 30 keywords
//
// Test cases:
//   1. Short spam (30 chars) - minimal tokenization
//   2. Medium legitimate (200 chars) - typical post
//   3. Long spam (1000 chars) - high tokenization/regex cost
//   4. MaxInputLength content (4096 chars) - truncation path
//   5. Early exit comparison - same content, full vs early exit

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
	println("=== GAS BENCHMARK: FULL PIPELINE ===\n")

	// ---- Setup: populate all state to realistic sizes

	corpus := antispam.NewCorpus()
	fps := antispam.NewFingerprintStore()
	bl := antispam.NewBlocklist()
	dict := antispam.NewKeywordDict()

	// Corpus: balanced spam + ham training
	spamSamples := []string{
		"free airdrop bonus click claim tokens prize money",
		"exclusive limited offer casino jackpot gambling bet",
		"earn income passive guaranteed investment returns profit",
		"click here free prize bonus earn money fast now",
		"claim free airdrop bonus tokens exclusive offer deal",
		"casino jackpot win prize free money bonus gambling",
		"investment guaranteed profit returns earn passive income",
		"free giveaway contest winner selected congratulations",
		"urgent action required claim prize money bonus now",
		"limited exclusive deal offer discount sale free bonus",
	}
	hamSamples := []string{
		"governance proposal voting delegation validator commission",
		"protocol upgrade implementation consensus mechanism design",
		"smart contract deployment transaction execution state",
		"validator delegation reward commission staking pool rate",
		"development update release version changelog patch notes",
		"community discussion forum feedback improvement suggestion",
		"documentation tutorial guide getting started reference",
		"review pull request changes code implementation test",
		"architecture design pattern interface implementation module",
		"debugging issue investigation root cause analysis fix",
	}
	for i := 0; i < 5; i++ {
		for _, s := range spamSamples {
			corpus.Train(s, true)
		}
		for _, s := range hamSamples {
			corpus.Train(s, false)
		}
	}

	// Fingerprints: 50 known spam entries
	for i := 0; i < 50; i++ {
		fps.Add("Spam variation number " + intToString(i) + " with unique filler to avoid exact overlap between entries")
	}

	// Blocklist: 5 patterns + 10 blocked addresses
	bl.AddPattern(`(?i)send\s+\d+\s+\w+\s+get\s+\d+`)
	bl.AddPattern(`(?i)free\s+airdrop`)
	bl.AddPattern(`(?i)click\s+here`)
	bl.AddPattern(`(?i)guaranteed\s+returns`)
	bl.AddPattern(`(?i)double\s+your`)
	for i := 0; i < 10; i++ {
		bl.BlockAddress("g1blocked" + intToString(i))
	}

	// Keywords: 30 weighted entries
	dict.BulkAdd(`free:3
airdrop:3
bonus:3
prize:3
casino:3
jackpot:3
click:2
claim:2
limited:2
exclusive:2
offer:2
earn:2
profit:2
guaranteed:2
investment:2
returns:2
passive:2
money:2
fast:2
urgent:2
contest:2
winner:2
giveaway:2
discount:2
deal:2
selected:1
congratulations:1
gift:1
reward:1
hurry:1`)

	println("State: corpus=" + intToString(corpus.Size()) + " tokens, fps=" + intToString(fps.Size()) + " entries, keywords=" + intToString(dict.Size()) + " entries")
	println("")

	goodRep := antispam.ReputationData{
		AccountAgeDays: 100,
		Balance:        5000000000,
		HasUsername:    true,
		TotalPosts:     50,
	}
	normalRate := antispam.RateState{PostCount: 3, WindowSeconds: 3600}

	// ---- TEST 1: Short Spam (30 chars)

	println("TEST 1: Short Spam (30 chars)")
	println("----")

	shortSpam := "Free airdrop! Claim now!!!"

	result1 := antispam.Score(antispam.ScoreInput{
		Author:  "g1user1",
		Content: shortSpam,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Content:   \"" + shortSpam + "\"")
	println("Score:     " + intToString(result1.Total))
	println("Triggered: " + sliceToString(result1.Triggered))
	println("")

	// ---- TEST 2: Medium Legitimate (200 chars)

	println("TEST 2: Medium Legitimate (200 chars)")
	println("----")

	mediumLegit := "The governance proposal for increasing validator delegation limits has been submitted for community review. Voting starts tomorrow and runs for 7 days. Please review the technical spec before voting."

	result2 := antispam.Score(antispam.ScoreInput{
		Author:  "g1user2",
		Content: mediumLegit,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Content:   (200 chars, governance discussion)")
	println("Score:     " + intToString(result2.Total))
	println("Triggered: " + sliceToString(result2.Triggered))
	println("")

	// ---- TEST 3: Long Spam (repeating pattern)

	println("TEST 3: Long Spam (~1000 chars)")
	println("----")

	segment := "Free bonus airdrop claim prize exclusive limited offer click here now! "
	longSpam := ""
	for len(longSpam) < 1000 {
		longSpam += segment
	}

	result3 := antispam.Score(antispam.ScoreInput{
		Author:  "g1user3",
		Content: longSpam,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Content:   (~1000 chars, repetitive spam keywords)")
	println("Score:     " + intToString(result3.Total))
	println("Triggered: " + sliceToString(result3.Triggered))
	println("")

	// ---- TEST 4: MaxInputLength (4096 chars, truncation path)

	println("TEST 4: MaxInputLength Content (4096 chars)")
	println("----")

	filler := "This is a long message that exercises the MaxInputLength truncation path in the scoring function to ensure gas costs are bounded even for oversized inputs. "
	maxContent := ""
	for len(maxContent) < 4200 {
		maxContent += filler
	}

	result4 := antispam.Score(antispam.ScoreInput{
		Author:  "g1user4",
		Content: maxContent,
		Rate:    normalRate,
		Rep:     goodRep,
		Corpus:  corpus,
		Fps:     fps,
		Bl:      bl,
		Dict:    dict,
	})

	println("Content:   (4096+ chars, truncated at MaxInputLength)")
	println("Score:     " + intToString(result4.Total))
	println("Triggered: " + sliceToString(result4.Triggered))
	println("")

	// ---- TEST 5: Early Exit Comparison

	println("TEST 5: Early Exit Comparison")
	println("----")

	burstRate := antispam.RateState{PostCount: 25, WindowSeconds: 3600}
	badRep := antispam.ReputationData{
		AccountAgeDays: 0,
		Balance:        100000000,
		HasUsername:    false,
		BanCount:       1,
	}

	result5full := antispam.Score(antispam.ScoreInput{
		Author:      "g1spammer",
		Content:     shortSpam,
		Rate:        burstRate,
		Rep:         badRep,
		Corpus:      corpus,
		Fps:         fps,
		Bl:          bl,
		Dict:        dict,
		EarlyExitAt: antispam.EarlyExitDisabled,
	})

	result5early := antispam.Score(antispam.ScoreInput{
		Author:      "g1spammer",
		Content:     shortSpam,
		Rate:        burstRate,
		Rep:         badRep,
		Corpus:      corpus,
		Fps:         fps,
		Bl:          bl,
		Dict:        dict,
		EarlyExitAt: antispam.ThresholdReject,
	})

	println("Full scoring:  score=" + intToString(result5full.Total) + " rules=" + sliceToString(result5full.Triggered))
	println("Early exit:    score=" + intToString(result5early.Total) + " rules=" + sliceToString(result5early.Triggered))
	println("Rules skipped: " + intToString(len(result5full.Triggered)-len(result5early.Triggered)))
	println("")

	// ---- SUMMARY

	println("SUMMARY")
	println("----")
	println("Test  Content        Score  Rules")
	println("1     Short spam     " + intToString(result1.Total) + "      " + sliceToString(result1.Triggered))
	println("2     Medium legit   " + intToString(result2.Total) + "      " + sliceToString(result2.Triggered))
	println("3     Long spam      " + intToString(result3.Total) + "      " + sliceToString(result3.Triggered))
	println("4     Max length     " + intToString(result4.Total) + "      " + sliceToString(result4.Triggered))
	println("5a    Full scoring   " + intToString(result5full.Total) + "      " + sliceToString(result5full.Triggered))
	println("5b    Early exit     " + intToString(result5early.Total) + "      " + sliceToString(result5early.Triggered))
}

// Output:
// === GAS BENCHMARK: FULL PIPELINE ===
//
// State: corpus=95 tokens, fps=50 entries, keywords=30 entries
//
// TEST 1: Short Spam (30 chars)
// ----
// Content:   "Free airdrop! Claim now!!!"
// Score:     11
// Triggered: [BLOCKED_PATTERN BAYES_SPAM KEYWORD_SPAM]
//
// TEST 2: Medium Legitimate (200 chars)
// ----
// Content:   (200 chars, governance discussion)
// Score:     0
// Triggered: []
//
// TEST 3: Long Spam (~1000 chars)
// ----
// Content:   (~1000 chars, repetitive spam keywords)
// Score:     11
// Triggered: [BLOCKED_PATTERN BAYES_SPAM KEYWORD_SPAM]
//
// TEST 4: MaxInputLength Content (4096 chars)
// ----
// Content:   (4096+ chars, truncated at MaxInputLength)
// Score:     0
// Triggered: []
//
// TEST 5: Early Exit Comparison
// ----
// Full scoring:  score=20 rules=[RATE_BURST NEW_ACCOUNT NO_USERNAME LOW_BALANCE BANNED_BEFORE BLOCKED_PATTERN BAYES_SPAM KEYWORD_SPAM]
// Early exit:    score=9 rules=[RATE_BURST NEW_ACCOUNT NO_USERNAME LOW_BALANCE BANNED_BEFORE]
// Rules skipped: 3
//
// SUMMARY
// ----
// Test  Content        Score  Rules
// 1     Short spam     11      [BLOCKED_PATTERN BAYES_SPAM KEYWORD_SPAM]
// 2     Medium legit   0      []
// 3     Long spam      11      [BLOCKED_PATTERN BAYES_SPAM KEYWORD_SPAM]
// 4     Max length     0      []
// 5a    Full scoring   20      [RATE_BURST NEW_ACCOUNT NO_USERNAME LOW_BALANCE BANNED_BEFORE BLOCKED_PATTERN BAYES_SPAM KEYWORD_SPAM]
// 5b    Early exit     9      [RATE_BURST NEW_ACCOUNT NO_USERNAME LOW_BALANCE BANNED_BEFORE]
