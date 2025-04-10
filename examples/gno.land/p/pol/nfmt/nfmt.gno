package nfmt

import (
	"errors"
	"strconv"

	// "reflect"
	// "encoding/binary"
	// "math"

	"gno.land/p/demo/ufmt"
	"gno.land/p/pol/unsafe"
)

func formatFromFloat(n float64) string {
	suffixes := []struct {
		value 	float64
		suffix 	string
	} {
		{1e24, "Y"},     // yotta
		{1e21, "Z"},     // zetta
		{1e18, "E"},     // exa
		{1e15, "P"},     // peta
		{1e12, "Tn"},    // trillion
		{1e9,  "Bn"},    // billion
		{1e6,  "M"},     // million
		{1e3,  "k"},     // thousand
	}

	for _, s := range suffixes {
		if n >= s.value {
			return ufmt.Sprintf("%.2f%s", n/s.value, s.suffix)
		}
	}
	return ufmt.Sprintf("%.0f", n)
}

//
// Convert any number into a formated number.
//	
// Examples:
//
// * Format("3000") -> "3k"
//
// * Format(3_980_000) -> "3.98M"
//
//
func Format(n any) (string, error) {
	switch v := n.(type) {
	case float64:
		return formatFromFloat(v), nil
	case float32:
		return formatFromFloat(float64(v)), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		x, err := unsafe.ObjectToBits(v)
		if err != nil {
			return "", err
		}
		return formatFromFloat(float64(x)), nil
	case string:
		x, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return "", err
		}
		return formatFromFloat(x), nil
	case interface { String() string }:
		return Format(v.String())
	}
	return "", errors.New("invalid type")
}