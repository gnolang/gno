package chainmap

import (
	"math"
	"testing"
)

func TestSquarifyPreservesAreasAndStaysInside(t *testing.T) {
	t.Parallel()

	frame := rect{0, 0, 1.6, 1}
	weights := []float64{40, 12, 9, 9, 5, 3, 1, 1, 1}
	got := squarify(weights, frame)
	if len(got) != len(weights) {
		t.Fatalf("rects = %d, want %d", len(got), len(weights))
	}

	var total, covered float64
	for _, w := range weights {
		total += w
	}
	const eps = 1e-9
	for i, r := range got {
		want := weights[i] / total * frame.W * frame.H
		if math.Abs(r.W*r.H-want) > 1e-6 {
			t.Errorf("rect %d area = %f, want %f", i, r.W*r.H, want)
		}
		if r.X < -eps || r.Y < -eps || r.X+r.W > frame.W+eps || r.Y+r.H > frame.H+eps {
			t.Errorf("rect %d = %+v leaves the frame", i, r)
		}
		covered += r.W * r.H
	}
	if math.Abs(covered-frame.W*frame.H) > 1e-6 {
		t.Errorf("rects cover %f, want the whole frame %f", covered, frame.W*frame.H)
	}

	for i := range got {
		for j := i + 1; j < len(got); j++ {
			if overlap(got[i], got[j]) > 1e-9 {
				t.Errorf("rects %d and %d overlap: %+v %+v", i, j, got[i], got[j])
			}
		}
	}
}

// Equal weights are the no-indexer map: one tile per package. They should come
// out close to square, not as strips.
func TestSquarifyKeepsEqualTilesSquareish(t *testing.T) {
	t.Parallel()

	weights := make([]float64, 36)
	for i := range weights {
		weights[i] = 1
	}
	for i, r := range squarify(weights, rect{0, 0, 1, 1}) {
		if ratio := max(r.W/r.H, r.H/r.W); ratio > 2 {
			t.Errorf("tile %d aspect ratio = %.2f, want at most 2", i, ratio)
		}
	}
}

func TestSquarifyDegenerateInput(t *testing.T) {
	t.Parallel()

	if got := squarify(nil, rect{0, 0, 1, 1}); len(got) != 0 {
		t.Errorf("no weights gave %d rects", len(got))
	}
	if got := squarify([]float64{1, 1}, rect{0, 0, 0, 1}); len(got) != 2 {
		t.Errorf("a zero-width frame must still give one rect per weight, got %d", len(got))
	}
}

func TestPercentIsRelativeToTheFrame(t *testing.T) {
	t.Parallel()

	frame := rect{0.4, 0.2, 0.8, 0.5}
	got := percent(rect{0.6, 0.45, 0.2, 0.25}, frame)
	want := Box{Left: 25, Top: 50, Width: 25, Height: 50}
	if got != want {
		t.Errorf("percent = %+v, want %+v", got, want)
	}
}

func overlap(a, b rect) float64 {
	w := min(a.X+a.W, b.X+b.W) - max(a.X, b.X)
	h := min(a.Y+a.H, b.Y+b.H) - max(a.Y, b.Y)
	if w <= 0 || h <= 0 {
		return 0
	}
	return w * h
}
