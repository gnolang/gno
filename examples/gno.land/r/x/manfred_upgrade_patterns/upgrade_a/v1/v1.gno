package upgradea

import (
	"strconv"
)

var counter int

func Inc(cur realm, nb int) {
	counter += nb
}

func GetInt() int {
	return counter
}

func GetString() string {
	return "v1 - " + strconv.Itoa(counter)
}
