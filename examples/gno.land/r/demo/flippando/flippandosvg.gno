package flippando

import (
	"std"
	"time"
	"bytes"
	"strings"
	"strconv"
	"errors"
	"encoding/base64"

	"gno.land/p/demo/ufmt"
)


func GenerateColorSquareGrid(solvedBoard []int64, size int) (string, error) {
	svgSquare := "<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 100 100\">"
	for i := int64(0); i < int64(size); i++ {
		for j := int64(0); j < int64(size); j++ {
			index := i*int64(size) + j
			key := solvedBoard[index]
			color, _ := getColorFromKey(key)
			rect := ufmt.Sprintf(`<rect width="25" height="25" fill="%s" x="%s" y="%s" />`, color, intToString(j*25), intToString(i*25))
			svgSquare += rect
		}
	}
	svgSquare += `</svg>`
	base64Svg := base64.StdEncoding.EncodeToString([]byte(svgSquare))
	return base64Svg, nil
}

// util
func intToString(value int64) string {
	formatted := strconv.FormatUint(uint64(value), 10)
	return formatted;
}

func getColorFromKey(key int64) (string, error) {
    switch key {
    case int64(1):
        return "#D9362A", nil
    case int64(2):
        return "#BE185D", nil
    case int64(3):
        return "#EC1818", nil
    case int64(4):
        return "#F4AA24", nil
    case int64(5):
        return "#F4D424", nil
    case int64(6):
        return "#F1F223", nil
    case int64(7):
        return "#A5DD0C", nil
    case int64(8):
        return "#37B400", nil
    case int64(9):
        return "#98D0E9", nil
    case int64(10):
        return "#3131FD", nil
    case int64(11):
        return "#1D2C85", nil
    case int64(12):
        return "#F2F2F2", nil
    case int64(13):
        return "#1A1A1A", nil
    case int64(14):
        return "#71842F", nil
    case int64(15):
        return "#90572F", nil
    case int64(16):
        return "#8316C0", nil
    default:
        panic("Invalid color key")
    }
}
