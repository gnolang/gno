package flippandoserver

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


func GenerateColorSquareGrid(solvedBoard []uint8, size uint) (string, error) {
	svgSquare := "<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 100 100\">"
	for i := uint(0); i < size; i++ {
		for j := uint(0); j < size; j++ {
			index := i*size + j
			key := solvedBoard[index]
			color, _ := getColorFromKey(key)
			rect := ufmt.Sprintf(`<rect width="25" height="25" fill="%s" x="%s" y="%s" />`, color, uintToString(j*25), uintToString(i*25))
			svgSquare += rect
		}
	}
	svgSquare += `</svg>`
	base64Svg := base64.StdEncoding.EncodeToString([]byte(svgSquare))
	return base64Svg, nil
}

// util
func uintToString(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}

func getColorFromKey(key uint8) (string, error) {
    switch key {
    case 1:
        return "#D9362A", nil
    case 2:
        return "#BE185D", nil
    case 3:
        return "#EC1818", nil
    case 4:
        return "#F4AA24", nil
    case 5:
        return "#F4D424", nil
    case 6:
        return "#F1F223", nil
    case 7:
        return "#A5DD0C", nil
    case 8:
        return "#37B400", nil
    case 9:
        return "#98D0E9", nil
    case 10:
        return "#3131FD", nil
    case 11:
        return "#1D2C85", nil
    case 12:
        return "#F2F2F2", nil
    case 13:
        return "#1A1A1A", nil
    case 14:
        return "#71842F", nil
    case 15:
        return "#90572F", nil
    case 16:
        return "#8316C0", nil
    default:
        panic("Invalid color key")
    }
}
