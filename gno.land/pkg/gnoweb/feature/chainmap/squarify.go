package chainmap

// rect is a box in map units, as squarify works on it.
type rect struct{ X, Y, W, H float64 }

// squarify lays weights out in r, one rect per weight in the same order, each
// rect's area proportional to its weight. weights must be sorted in
// descending order.
//
// This is the squarified treemap of Bruls, Huizing and van Wijk: items fill a
// row along the shorter side of what is left for as long as adding one does
// not worsen the row's most elongated rect. Squares read as tiles; a plain
// slice-and-dice layout turns a namespace with many packages into slivers.
func squarify(weights []float64, r rect) []rect {
	out := make([]rect, 0, len(weights))
	var total float64
	for _, w := range weights {
		total += w
	}
	if total <= 0 || r.W <= 0 || r.H <= 0 {
		return make([]rect, len(weights))
	}

	areas := make([]float64, len(weights))
	for i, w := range weights {
		areas[i] = w * r.W * r.H / total
	}

	for i := 0; i < len(areas); {
		side := min(r.W, r.H)
		j := i + 1
		for j < len(areas) && worst(areas[i:j+1], side) <= worst(areas[i:j], side) {
			j++
		}
		row := areas[i:j]
		var rowArea float64
		for _, a := range row {
			rowArea += a
		}

		if r.W >= r.H {
			// A column against the left edge, as tall as what is left.
			colW := rowArea / r.H
			y := r.Y
			for _, a := range row {
				h := a / colW
				out = append(out, rect{r.X, y, colW, h})
				y += h
			}
			r.X += colW
			r.W -= colW
		} else {
			// A row against the top edge, as wide as what is left.
			rowH := rowArea / r.W
			x := r.X
			for _, a := range row {
				w := a / rowH
				out = append(out, rect{x, r.Y, w, rowH})
				x += w
			}
			r.Y += rowH
			r.H -= rowH
		}
		i = j
	}
	return out
}

// worst is the largest aspect ratio a row of these areas would have when laid
// along a side of the given length.
func worst(row []float64, side float64) float64 {
	var sum float64
	hi, lo := row[0], row[0]
	for _, a := range row {
		sum += a
		hi = max(hi, a)
		lo = min(lo, a)
	}
	s2, w2 := sum*sum, side*side
	return max(w2*hi/s2, s2/(w2*lo))
}
