package antispam

const (
	// guardDefaultMinScore is the minimum total spam score for auto-training.
	// Set above ThresholdReject (8) so only unambiguous spam is auto-trained.
	// Borderline content (score 8-9) is rejected but NOT fed to the corpus.
	guardDefaultMinScore = 10

	// guardDefaultMinRules is the minimum number of distinct triggered rules
	// for auto-training. Multi-rule consensus prevents single-rule bias
	// from poisoning the corpus (e.g. BAYES_SPAM alone should not self-reinforce).
	guardDefaultMinRules = 3

	// guardDefaultMaxTrains is the circuit breaker limit. After this many
	// auto-trains, the guard stops allowing training until an admin calls
	// Reset() after reviewing the corpus state.
	guardDefaultMaxTrains = 500
)

// TrainingGuard enforces safety constraints on auto-training the Corpus.
// Prevents feedback loops where borderline content gets auto-trained as
// spam, biasing the corpus to flag similar legitimate content.
//
// Three safeguards:
//  1. Minimum score - only clear spam, not borderline (default: 10 > ThresholdReject)
//  2. Multi-rule consensus - multiple rules must fire (default: 3)
//  3. Circuit breaker - cap total auto-trains until admin review (default: 500)
//
// Usage:
//
//	var guard = antispam.NewTrainingGuard()
//
//	func ModeratePost(author, content string) {
//	    score := antispam.Score(...)
//	    if score.Total >= ThresholdReject {
//	        if guard.ShouldTrain(score) {
//	            corpus.Train(content, true)
//	            guard.RecordTrain()
//	        }
//	    }
//	}
type TrainingGuard struct {
	minScore   int
	minRules   int
	maxTrains  int
	trainCount int
	tripped    bool
}

// NewTrainingGuard creates a TrainingGuard with sensible defaults.
func NewTrainingGuard() *TrainingGuard {
	return &TrainingGuard{
		minScore:  guardDefaultMinScore,
		minRules:  guardDefaultMinRules,
		maxTrains: guardDefaultMaxTrains,
	}
}

// NewTrainingGuardCustom creates a TrainingGuard with custom thresholds.
// Zero or negative values are replaced with defaults.
func NewTrainingGuardCustom(minScore, minRules, maxTrains int) *TrainingGuard {
	if minScore <= 0 {
		minScore = guardDefaultMinScore
	}
	if minRules <= 0 {
		minRules = guardDefaultMinRules
	}
	if maxTrains <= 0 {
		maxTrains = guardDefaultMaxTrains
	}
	return &TrainingGuard{
		minScore:  minScore,
		minRules:  minRules,
		maxTrains: maxTrains,
	}
}

// ShouldTrain returns true if the spam score meets safety thresholds
// and the circuit breaker has not tripped. Checks (in order):
//  1. Circuit breaker not tripped
//  2. Score >= minScore
//  3. Number of triggered rules >= minRules
func (g *TrainingGuard) ShouldTrain(score SpamScore) bool {
	if g.tripped || g.minScore <= 0 {
		return false
	}
	if score.Total < g.minScore {
		return false
	}
	if len(score.Triggered) < g.minRules {
		return false
	}
	return true
}

// RecordTrain increments the training counter and trips the circuit
// breaker when maxTrains is reached. Call after every auto-train.
func (g *TrainingGuard) RecordTrain() {
	g.trainCount++
	if g.trainCount >= g.maxTrains {
		g.tripped = true
	}
}

// Reset clears the training counter and untrips the circuit breaker.
// Intended for admin use after reviewing corpus state.
func (g *TrainingGuard) Reset() {
	g.trainCount = 0
	g.tripped = false
}

// IsTripped returns true if the circuit breaker has been activated.
func (g *TrainingGuard) IsTripped() bool {
	return g.tripped
}

// TrainCount returns the number of auto-trains since the last reset.
func (g *TrainingGuard) TrainCount() int {
	return g.trainCount
}
