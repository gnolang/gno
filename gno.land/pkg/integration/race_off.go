//go:build !race

package integration

// raceEnabled is false when the binary is compiled without -race.
const raceEnabled = false
