package upgradea

import (
	"strconv"

	v1 "gno.land/r/x/manfred_upgrade_patterns/upgrade_a/v1"
)

var counter int

func Inc(nb int) {
	crossing()

	counter += nb
}

func GetInt() int {
	v1Counter := v1.GetInt()
	sum := v1Counter + counter
	return sum
}

func GetString() string {
	val := GetInt()
	return "v2 - " + strconv.Itoa(val)
}
