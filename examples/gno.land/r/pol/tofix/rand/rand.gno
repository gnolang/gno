package rand

import (
	"strconv"
	"math/rand"
)

func Render(_ string) string {
	return strconv.Itoa(rand.Int()) // this should crash
}