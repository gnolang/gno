package main

import "math"

type pair struct {
	a uint256
	b int256
}

func main() {
	m := map[uint256]string{}
	m[0] = "zero"
	m[1] = "one"
	m[math.MaxUint256] = "max"
	m[1<<255] = "half"
	var z uint256
	println(len(m), m[z], m[1], m[math.MaxUint256], m[1<<255])
	delete(m, 1)
	_, ok := m[1]
	println(len(m), ok)

	// int256 and uint256 keys with the same bits are distinct.
	mi := map[any]int{}
	mi[uint256(1)] = 1
	mi[int256(1)] = 2
	mi[int256(-1)] = 3
	mi[uint256(math.MaxUint256)] = 4
	println(len(mi), mi[uint256(1)], mi[int256(1)], mi[int256(-1)], mi[uint256(math.MaxUint256)])

	// In composite keys.
	mp := map[pair]bool{}
	mp[pair{1, -1}] = true
	mp[pair{1, -1}] = true
	mp[pair{math.MaxUint256, math.MinInt256}] = true
	println(len(mp), mp[pair{1, -1}], mp[pair{1, 1}])
	ma := map[[2]int256]int{{-1, 1}: 1}
	println(ma[[2]int256{-1, 1}], ma[[2]int256{1, -1}])
}

// Output:
// 4 zero one max half
// 3 false
// 4 1 2 3 4
// 2 true false
// 1 0
