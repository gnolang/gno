package antispam

import (
	"testing"
)

// ---- Crypto content false positive tests ----
//
// Validates that legitimate Gno ecosystem content (governance, staking,
// token economics, airdrops) is not flagged as spam when using a keyword
// dictionary containing dual-use blockchain terminology.

// cryptoTestSetup builds shared state for all crypto content tests.
// Uses a keyword dictionary with terms that overlap between spam and
// legitimate blockchain usage, plus a balanced corpus.
func cryptoTestSetup() (*Corpus, *KeywordDict) {
	// Keyword dictionary with dual-use crypto terms.
	// These are realistic weights an admin might configure.
	dict := NewKeywordDict()
	dict.BulkAdd(`airdrop:3
mint:2
claim:2
whitelist:2
token:1
free:3
bonus:3
prize:3
exclusive:2
limited:2
earn:2
casino:3
jackpot:3
click:2
money:2`)

	// Balanced corpus: spam + legitimate crypto content.
	corpus := NewCorpus()

	// Spam training
	corpus.Train("free airdrop bonus click claim tokens prize money", true)
	corpus.Train("exclusive limited offer casino jackpot free bonus", true)
	corpus.Train("click here free prize bonus earn money fast", true)
	corpus.Train("claim free airdrop bonus tokens exclusive offer", true)
	corpus.Train("free money earn bonus prize jackpot casino now", true)

	// Legitimate crypto training (governance, staking, dev)
	corpus.Train("governance proposal voting delegation validator commission", false)
	corpus.Train("staking rewards delegation validator commission rate", false)
	corpus.Train("token economics supply distribution emission schedule", false)
	corpus.Train("proposal discussion governance community voting period", false)
	corpus.Train("validator delegation reward commission staking pool", false)
	corpus.Train("mint tokens burn supply inflation deflationary model", false)
	corpus.Train("airdrop distribution token allocation community governance", false)
	corpus.Train("claim rewards staking delegation validator earnings", false)

	return corpus, dict
}

func TestCryptoContentFalsePositives(t *testing.T) {
	corpus, dict := cryptoTestSetup()

	// Good reputation: established community member.
	// Rules out reputation-based penalties as noise.
	goodRep := ReputationData{
		AccountAgeDays: 180,
		Balance:        5000000000, // 5000 GNOT
		HasUsername:    true,
		TotalAccepted:  50,
	}
	normalRate := RateState{PostCount: 2, WindowSeconds: 3600}

	t.Run("legit: governance proposal discussion", func(t *testing.T) {
		content := "The governance proposal for increasing validator delegation limits has been submitted. Community voting starts tomorrow and runs for 7 days. Please review the changes before casting your vote."

		result := Score(ScoreInput{
			Author:  "g1contributor",
			Content: content,
			Rate:    normalRate,
			Rep:     goodRep,
			Corpus:  corpus,
			Dict:    dict,
		})

		if result.Total >= ThresholdHide {
			t.Errorf("FALSE POSITIVE: governance content scored %d (threshold=%d). Triggered: %v",
				result.Total, ThresholdHide, result.Triggered)
		}
	})

	t.Run("legit: staking rewards discussion", func(t *testing.T) {
		content := "To claim your staking rewards, delegate to a validator and earn a commission based on their performance. Rewards are distributed proportionally to your stake amount."

		result := Score(ScoreInput{
			Author:  "g1validator",
			Content: content,
			Rate:    normalRate,
			Rep:     goodRep,
			Corpus:  corpus,
			Dict:    dict,
		})

		if result.Total >= ThresholdHide {
			t.Errorf("FALSE POSITIVE: staking content scored %d (threshold=%d). Triggered: %v",
				result.Total, ThresholdHide, result.Triggered)
		}
	})

	t.Run("legit: token economics with mint and token", func(t *testing.T) {
		content := "The proposed token emission schedule would mint 1000 tokens per block initially, with a halving every year. This deflationary model balances supply growth with long-term scarcity."

		result := Score(ScoreInput{
			Author:  "g1economist",
			Content: content,
			Rate:    normalRate,
			Rep:     goodRep,
			Corpus:  corpus,
			Dict:    dict,
		})

		if result.Total >= ThresholdHide {
			t.Errorf("FALSE POSITIVE: tokenomics content scored %d (threshold=%d). Triggered: %v",
				result.Total, ThresholdHide, result.Triggered)
		}
	})

	t.Run("legit: airdrop announcement from core team", func(t *testing.T) {
		content := "The governance airdrop for early contributors will distribute tokens based on historical participation. Eligible addresses were snapshotted at block 500000. The distribution rationale is published on the forum."

		result := Score(ScoreInput{
			Author:  "g1coreteam",
			Content: content,
			Rate:    normalRate,
			Rep:     goodRep,
			Corpus:  corpus,
			Dict:    dict,
		})

		if result.Total >= ThresholdHide {
			t.Errorf("FALSE POSITIVE: airdrop announcement scored %d (threshold=%d). Triggered: %v",
				result.Total, ThresholdHide, result.Triggered)
		}
	})

	t.Run("spam: obvious scam with crypto keywords", func(t *testing.T) {
		content := "FREE AIRDROP!!! Claim exclusive bonus tokens NOW! Limited whitelist spots! Mint 1000 tokens FREE! Click here for prize!!!"

		result := Score(ScoreInput{
			Author:  "g1scammer",
			Content: content,
			Rate:    RateState{PostCount: 15, WindowSeconds: 3600},
			Rep: ReputationData{
				AccountAgeDays: 0,
				Balance:        100000000, // 100 GNOT (low)
				HasUsername:    false,
			},
			Corpus: corpus,
			Dict:   dict,
		})

		if result.Total < ThresholdReject {
			t.Errorf("MISSED SPAM: obvious scam scored only %d (threshold=%d). Triggered: %v",
				result.Total, ThresholdReject, result.Triggered)
		}
	})

	t.Run("borderline: short post with multiple crypto keywords", func(t *testing.T) {
		// Documents behavior of short posts with dual-use keywords.
		// If this triggers ThresholdHide, the keyword weights or
		// co-occurrence thresholds may need tuning for crypto realms.
		content := "Mint new tokens then stake them to earn validator rewards."

		result := Score(ScoreInput{
			Author:  "g1user",
			Content: content,
			Rate:    normalRate,
			Rep:     goodRep,
			Corpus:  corpus,
			Dict:    dict,
		})

		// Log for awareness regardless of outcome.
		t.Logf("Borderline crypto post: score=%d, triggered=%v", result.Total, result.Triggered)

		if result.Total >= ThresholdHide {
			t.Errorf("FALSE POSITIVE: borderline crypto post scored %d (threshold=%d). Triggered: %v. Consider adjusting keyword weights for crypto realms.",
				result.Total, ThresholdHide, result.Triggered)
		}
	})
}
