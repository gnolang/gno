package antispam

import (
	"testing"
)

// Constructor tests

func TestNewTrainingGuard(t *testing.T) {
	t.Run("defaults: correct threshold values", func(t *testing.T) {
		g := NewTrainingGuard()
		if g.minScore != guardDefaultMinScore {
			t.Errorf("expected minScore=%d, got %d", guardDefaultMinScore, g.minScore)
		}
		if g.minRules != guardDefaultMinRules {
			t.Errorf("expected minRules=%d, got %d", guardDefaultMinRules, g.minRules)
		}
		if g.maxTrains != guardDefaultMaxTrains {
			t.Errorf("expected maxTrains=%d, got %d", guardDefaultMaxTrains, g.maxTrains)
		}
		if g.trainCount != 0 {
			t.Errorf("expected trainCount=0, got %d", g.trainCount)
		}
		if g.tripped {
			t.Error("expected tripped=false")
		}
	})
}

func TestNewTrainingGuardCustom(t *testing.T) {
	t.Run("custom: accepts positive values", func(t *testing.T) {
		g := NewTrainingGuardCustom(15, 5, 1000)
		if g.minScore != 15 {
			t.Errorf("expected minScore=15, got %d", g.minScore)
		}
		if g.minRules != 5 {
			t.Errorf("expected minRules=5, got %d", g.minRules)
		}
		if g.maxTrains != 1000 {
			t.Errorf("expected maxTrains=1000, got %d", g.maxTrains)
		}
	})

	t.Run("custom: zero values use defaults", func(t *testing.T) {
		g := NewTrainingGuardCustom(0, 0, 0)
		if g.minScore != guardDefaultMinScore {
			t.Errorf("expected minScore=%d for zero input, got %d", guardDefaultMinScore, g.minScore)
		}
		if g.minRules != guardDefaultMinRules {
			t.Errorf("expected minRules=%d for zero input, got %d", guardDefaultMinRules, g.minRules)
		}
		if g.maxTrains != guardDefaultMaxTrains {
			t.Errorf("expected maxTrains=%d for zero input, got %d", guardDefaultMaxTrains, g.maxTrains)
		}
	})

	t.Run("custom: negative values use defaults", func(t *testing.T) {
		g := NewTrainingGuardCustom(-5, -1, -100)
		if g.minScore != guardDefaultMinScore {
			t.Errorf("expected minScore=%d for negative input, got %d", guardDefaultMinScore, g.minScore)
		}
		if g.minRules != guardDefaultMinRules {
			t.Errorf("expected minRules=%d for negative input, got %d", guardDefaultMinRules, g.minRules)
		}
		if g.maxTrains != guardDefaultMaxTrains {
			t.Errorf("expected maxTrains=%d for negative input, got %d", guardDefaultMaxTrains, g.maxTrains)
		}
	})
}

// ShouldTrain tests

func TestShouldTrain(t *testing.T) {
	t.Run("success: high score with multiple rules", func(t *testing.T) {
		g := NewTrainingGuard()
		score := SpamScore{
			Total:     12,
			Triggered: []string{RuleBayesSpam, RuleKeywordSpam, RuleAllCaps, RuleRateBurst},
		}
		if !g.ShouldTrain(score) {
			t.Error("expected ShouldTrain=true for score=12 with 4 rules")
		}
	})

	t.Run("reject: score below threshold", func(t *testing.T) {
		g := NewTrainingGuard()
		score := SpamScore{
			Total:     7,
			Triggered: []string{RuleBayesSpam, RuleKeywordSpam, RuleAllCaps},
		}
		if g.ShouldTrain(score) {
			t.Error("expected ShouldTrain=false for score=7 (below minScore=10)")
		}
	})

	t.Run("reject: too few rules triggered", func(t *testing.T) {
		g := NewTrainingGuard()
		score := SpamScore{
			Total:     12,
			Triggered: []string{RuleBayesSpam, RuleKeywordSpam},
		}
		if g.ShouldTrain(score) {
			t.Error("expected ShouldTrain=false for 2 rules (below minRules=3)")
		}
	})

	t.Run("reject: circuit breaker tripped", func(t *testing.T) {
		g := NewTrainingGuard()
		g.tripped = true
		score := SpamScore{
			Total:     15,
			Triggered: []string{RuleBayesSpam, RuleKeywordSpam, RuleAllCaps, RuleRateBurst},
		}
		if g.ShouldTrain(score) {
			t.Error("expected ShouldTrain=false when circuit breaker is tripped")
		}
	})

	t.Run("edge: exactly at minScore threshold (inclusive)", func(t *testing.T) {
		g := NewTrainingGuardCustom(10, 3, 500)
		score := SpamScore{
			Total:     10,
			Triggered: []string{"A", "B", "C"},
		}
		if !g.ShouldTrain(score) {
			t.Error("expected ShouldTrain=true at exact minScore (>= is inclusive)")
		}
	})

	t.Run("edge: exactly at minRules threshold (inclusive)", func(t *testing.T) {
		g := NewTrainingGuardCustom(10, 3, 500)
		score := SpamScore{
			Total:     15,
			Triggered: []string{"A", "B", "C"},
		}
		if !g.ShouldTrain(score) {
			t.Error("expected ShouldTrain=true at exact minRules (>= is inclusive)")
		}
	})

	t.Run("reject: empty triggered slice", func(t *testing.T) {
		g := NewTrainingGuard()
		score := SpamScore{
			Total:     15,
			Triggered: []string{},
		}
		if g.ShouldTrain(score) {
			t.Error("expected ShouldTrain=false for empty Triggered slice")
		}
	})
}

// RecordTrain and circuit breaker tests

func TestRecordTrain(t *testing.T) {
	t.Run("success: increments train count", func(t *testing.T) {
		g := NewTrainingGuard()
		if g.TrainCount() != 0 {
			t.Fatalf("expected initial count=0, got %d", g.TrainCount())
		}
		g.RecordTrain()
		if g.TrainCount() != 1 {
			t.Errorf("expected count=1, got %d", g.TrainCount())
		}
		g.RecordTrain()
		if g.TrainCount() != 2 {
			t.Errorf("expected count=2, got %d", g.TrainCount())
		}
	})

	t.Run("circuit: trips at maxTrains", func(t *testing.T) {
		g := NewTrainingGuardCustom(10, 3, 5)
		for i := 0; i < 4; i++ {
			g.RecordTrain()
			if g.IsTripped() {
				t.Errorf("unexpected trip at count %d", i+1)
			}
		}
		g.RecordTrain() // 5th train -> trips
		if !g.IsTripped() {
			t.Error("expected circuit to trip at maxTrains=5")
		}
	})

	t.Run("circuit: prevents training after trip", func(t *testing.T) {
		g := NewTrainingGuardCustom(5, 2, 3)
		score := SpamScore{Total: 10, Triggered: []string{"A", "B", "C"}}

		for i := 0; i < 3; i++ {
			if !g.ShouldTrain(score) {
				t.Errorf("expected ShouldTrain=true before trip (i=%d)", i)
			}
			g.RecordTrain()
		}

		if g.ShouldTrain(score) {
			t.Error("expected ShouldTrain=false after circuit trip")
		}
	})
}

// Reset tests

func TestTrainingGuardReset(t *testing.T) {
	t.Run("success: clears train count", func(t *testing.T) {
		g := NewTrainingGuard()
		g.RecordTrain()
		g.RecordTrain()
		g.RecordTrain()
		if g.TrainCount() != 3 {
			t.Fatalf("setup: expected count=3, got %d", g.TrainCount())
		}
		g.Reset()
		if g.TrainCount() != 0 {
			t.Error("expected count=0 after Reset")
		}
	})

	t.Run("success: untrips circuit breaker", func(t *testing.T) {
		g := NewTrainingGuardCustom(5, 2, 2)
		g.RecordTrain()
		g.RecordTrain()
		if !g.IsTripped() {
			t.Fatal("setup: expected tripped=true")
		}
		g.Reset()
		if g.IsTripped() {
			t.Error("expected tripped=false after Reset")
		}
	})

	t.Run("success: re-enables training after reset", func(t *testing.T) {
		g := NewTrainingGuardCustom(5, 2, 2)
		score := SpamScore{Total: 10, Triggered: []string{"A", "B", "C"}}

		g.RecordTrain()
		g.RecordTrain()
		if g.ShouldTrain(score) {
			t.Fatal("setup: expected tripped")
		}

		g.Reset()
		if !g.ShouldTrain(score) {
			t.Error("expected ShouldTrain=true after Reset")
		}
	})
}

// Zero-value edge case

func TestTrainingGuardZeroValue(t *testing.T) {
	t.Run("edge: zero-value guard rejects all", func(t *testing.T) {
		var g TrainingGuard
		score := SpamScore{Total: 100, Triggered: []string{"A", "B", "C", "D", "E"}}
		if g.ShouldTrain(score) {
			t.Error("zero-value TrainingGuard should reject all (minScore=0 means nothing passes)")
		}
	})
}
