package glicko2

import (
	"testing"
)

func TestExampleCalculations(t *testing.T) {
	// These are from the example in prof. Glickman's paper.
	// At the end, t.Log should print for the first player updated values
	// for Rating, RatingDeviation and RatingVolatility matching those in the
	// examples.
	ratings := []*PlayerRating{
		{ID: "1", Rating: 1500, RatingDeviation: 200, RatingVolatility: 0.06},
		{ID: "2", Rating: 1400, RatingDeviation: 30, RatingVolatility: 0.06},
		{ID: "3", Rating: 1550, RatingDeviation: 100, RatingVolatility: 0.06},
		{ID: "4", Rating: 1700, RatingDeviation: 300, RatingVolatility: 0.06},
	}
	scores := []RatingScore{
		{White: "1", Black: "2", Score: 1},
		{White: "1", Black: "3", Score: 0},
		{White: "1", Black: "4", Score: 0},
	}
	UpdateRatings(ratings, scores)
	r := ratings[0]
	t.Logf("%.4f (± %.4f, volatility: %.4f); working values: %.2f / %.2f / %.2f\n",
		r.Rating, r.RatingDeviation, r.RatingVolatility,
		r.wr, r.wrd, r.wrv,
	)
}
