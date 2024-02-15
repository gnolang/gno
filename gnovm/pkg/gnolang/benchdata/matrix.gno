// param: 3 4 5 6

package main

func main() {
	const p = {{ .Param }}
	for i := 0; i < {{ .N }}; i++ {
		_ = det(mul(dist(p), mul(id(p), dist(p))))
	}
}

// identity matrix
func id(sz int) [][]int {
	r := make([][]int, sz)
	for i := range r {
		r[i] = make([]int, sz)
		r[i][i] = 1
	}
	return r
}

// distance from corner
func dist(sz int) [][]int {
	half := sz / 2
	r := make([][]int, sz)
	for i := range r {
		r[i] = make([]int, sz)
		vdist := i
		if vdist >= half {
			vdist = sz - 1 - i
		}
		for j := range r[i] {
			hdist := j
			if hdist >= half {
				hdist = sz - 1 - j
			}
			r[i][j] = vdist + hdist
		}
	}
	return r
}

func det(m [][]int) int {
    size := len(m)
    if size == 2 {
        return m[0][0]*m[1][1] - m[0][1]*m[1][0]
    }
	subMatrix := make([][]int, size-1)
	for j := range subMatrix {
		subMatrix[j] = make([]int, size-1)
	}

    determinant := 0
    for i := 0; i < size; i++ {
        for j := 1; j < size; j++ {
            t := 0
            for k := 0; k < size; k++ {
                if k == i {
                    continue
                }
                subMatrix[j-1][t] = m[j][k]
                t++
            }
        }
		sign := 1
		if i % 2 == 1 {
			sign = -1
		}
        determinant += m[0][i] * det(subMatrix) * sign
    }
    return determinant
}

func mul(m1, m2 [][]int) [][]int {
	size := len(m1)
	result := make([][]int, size)
	for i := range result {
		result[i] = make([]int, size)
	}

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			for k := 0; k < size; k++ {
				result[i][j] += m1[i][k] * m2[k][j]
			}
		}
	}

	return result
}
