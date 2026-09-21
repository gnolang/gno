package glicko2

import (
	"math"
)

// http://www.glicko.net/glicko/glicko2.pdf

const (
	// Step 1.
	// Determine a rating and RD for each player at the onset of the rating period. The
	// system constant, τ , which constrains the change in volatility over time, needs to be
	// set prior to application of the system. Reasonable choices are between 0.3 and 1.2,
	// though the system should be tested to decide which value results in greatest predictive
	// accuracy. Smaller values of τ prevent the volatility measures from changing by large
	// amounts, which in turn prevent enormous changes in ratings based on very improbable
	// results. If the application of Glicko-2 is expected to involve extremely improbable
	// collections of game outcomes, then τ should be set to a small value, even as small as,
	// say, τ = 0.2.
	GlickoTau = 0.5

	GlickoInitialRating     = 1500
	GlickoInitialRD         = 350
	GlickoInitialVolatility = 0.06

	glickoScaleFactor = 173.7178
)

type PlayerRating struct {
	ID address

	Rating           float64
	RatingDeviation  float64
	RatingVolatility float64

	// working values, these are referred to as μ, φ and σ in the paper.
	wr, wrd, wrv float64
}

func NewPlayerRating(addr address) *PlayerRating {
	return &PlayerRating{
		ID:               addr,
		Rating:           GlickoInitialRating,
		RatingDeviation:  GlickoInitialRD,
		RatingVolatility: GlickoInitialVolatility,
	}
}

// RatingScore is the outcome of a game between two players.
type RatingScore struct {
	White, Black address
	Score        float64 // 0 = black win, 0.5 = draw, 1 = white win
}

func getRatingScore(scores []RatingScore, player, opponent address) float64 {
	for _, score := range scores {
		if score.White == player && score.Black == opponent {
			return score.Score
		}
		if score.Black == player && score.White == opponent {
			return 1 - score.Score
		}
	}
	return -1
}

func UpdateRatings(source []*PlayerRating, scores []RatingScore) {
	// step 2: assign working wr/wrd/wrv
	for _, player := range source {
		player.wr = (player.Rating - GlickoInitialRating) / glickoScaleFactor
		player.wrd = player.RatingDeviation / glickoScaleFactor
		player.wrv = player.RatingVolatility
	}

	for _, player := range source {
		// step 3: compute the quantity v. This is the estimated variance of the team’s/player’s
		// rating based only on game outcomes
		// step 4: compute the quantity ∆, the estimated improvement in rating
		v, delta, competed := glickoVDelta(player, source, scores)

		newRDPart := math.Sqrt(player.wrd*player.wrd + player.wrv*player.wrv)
		if !competed {
			// fast path for players who have not competed:
			// update only rating deviation
			player.RatingDeviation = newRDPart
			player.wr, player.wrd, player.wrv = 0, 0, 0
			continue
		}

		// step 5: determine the new value, σ′, of the volatility.
		player.wrv = glickoVolatility(player, delta, v)
		// step 6 and 7
		player.wrd = 1 / math.Sqrt(1/v+1/(newRDPart*newRDPart))
		player.wr = player.wr + player.wrd*player.wrd*(delta/v)

		// step 8
		player.Rating = player.wr*glickoScaleFactor + 1500
		player.RatingDeviation = player.wrd * glickoScaleFactor
		player.RatingVolatility = player.wrv
		player.wrv, player.wrd, player.wr = 0, 0, 0
	}
}

func glickoVDelta(player *PlayerRating, source []*PlayerRating, scores []RatingScore) (v, delta float64, competed bool) {
	var deltaPartial float64
	for _, opponent := range source {
		if opponent.ID == player.ID {
			continue
		}
		score := getRatingScore(scores, player.ID, opponent.ID)
		if score == -1 { // not found
			continue
		}
		competed = true

		// Step 3
		gTheta := 1 / math.Sqrt(1+(3*opponent.wrd*opponent.wrd)/(math.Pi*math.Pi))
		eMu := 1 / (1 + math.Exp(-gTheta*(player.wr-opponent.wr)))
		v += gTheta * gTheta * eMu * (1 - eMu)

		// step 4
		deltaPartial += gTheta * (score - eMu)
	}
	v = 1 / v
	return v, v * deltaPartial, competed
}

const glickoVolatilityEps = 0.000001

// Step 5
func glickoVolatility(player *PlayerRating, delta, v float64) float64 {
	aInit := math.Log(player.wrv * player.wrv)
	A := aInit
	var B float64
	rdp2 := player.wrd * player.wrd
	if delta*delta > rdp2+v {
		B = math.Log(delta*delta - rdp2 - v)
	} else {
		k := float64(1)
		for ; glickoVolatilityF(A-k*GlickoTau, aInit, rdp2, delta, v) < 0; k++ {
		}
		B = A - k*GlickoTau
	}
	fA, fB := glickoVolatilityF(A, aInit, rdp2, delta, v), glickoVolatilityF(B, aInit, rdp2, delta, v)
	for math.Abs(B-A) > glickoVolatilityEps {
		C := A + ((A-B)*fA)/(fB-fA)
		fC := glickoVolatilityF(C, aInit, rdp2, delta, v)
		if fC*fB <= 0 {
			A, fA = B, fB
		} else {
			fA = fA / 2
		}
		B, fB = C, fC
	}
	return math.Exp(A / 2)
}

// rdp2: player.rd, power 2
func glickoVolatilityF(x, a, rdp2, delta, v float64) float64 {
	expX := math.Exp(x)
	return (expX*(delta*delta-rdp2-v-expX))/(2*pow2(rdp2+v+expX)) -
		(x-a)/(GlickoTau*GlickoTau)
}

func pow2(f float64) float64 { return f * f }
