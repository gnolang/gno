package place

import (
	"encoding/base64"
	"testing"

	"gno.land/p/demo/ufmt"
)

var (
	testValidColor = []string{
		"rgb(255,0,0)",
		"rgb(0,0,0)",
		"rgb(255,255,255)",
	}

	testInvalidColor = []string{
		"rgb(255555,0,0)",
		"rgb(0,0)",
		"rgb(,0,0)",
		"rgb(1,0;0)",
		"rgb(1k,0,0)",
		"rgb(0,256,0)",
	}

	testSubmitInvalidCoordinates = []struct {
		x uint64
		y uint64
	}{
		{1000, 1},
		{1, 40},
	}

	testSubmitInvalidColor = []string{
		"This color is definitly not a valid one",
		"",
		"red",
		"Blue",
	}

	testValidAdress = []struct {
		address string
		x       uint64
		y       uint64
	}{
		{"someAdress/r/eva/place:?color=G&x=9&y=2", 9, 2},
		{"?x=19&y=1", 19, 1},
	}

	testInvalidAdress = []string{
		"someAdress/r/eva/place:?color=G&x=9a&y=2",
		"",
	}
)

func Test_generateSquareSvg_valid(t *testing.T) {
	value, err := generateSquareSvg("rgb(255,0,0)")

	if err != nil {
		t.Fatal("invalid color")
	}

	decodedData, err := base64.StdEncoding.DecodeString(value)

	correctValue := `<svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" viewBox=""><rect x="0" y="0" width="30" height="30" rx="0" ry="0" fill="rgb(255,0,0)" /></svg>`

	if err != nil {
		t.Fatal("Cannot decode the svg")
	}
	// Verify the svg is correct
	if string(decodedData) != correctValue {
		t.Fatalf("The svg created is invalid")
	}
}

func Test_generateSquareSvg_invalid(t *testing.T) {
	for _, testInput := range testInvalidColor {
		t.Run(testInput, func(t *testing.T) {
			_, err := generateSquareSvg(testInput)

			if err == nil {
				t.Fatal("color should be invalid")
			}
		})
	}
}

func Test_checkColorValid_true(t *testing.T) {
	for _, testInput := range testValidColor {
		t.Run(testInput, func(t *testing.T) {
			value := checkColorValid(testInput)

			// Verify the svg is correct
			if value == false {
				t.Fatalf("The color should be valid")
			}
		})
	}
}

func Test_checkColorValid_false(t *testing.T) {
	for _, testInput := range testInvalidColor {
		t.Run(testInput, func(t *testing.T) {
			value := checkColorValid(testInput)

			// Verify the svg is incorrect
			if value == true {
				t.Fatalf("The color should be invalid")
			}
		})
	}
}

func Test_submitInvalidCoordinates(t *testing.T) {
	for _, testInput := range testSubmitInvalidCoordinates {
		testName := ufmt.Sprintf("(%d,%d)", testInput.x, testInput.y)
		t.Run(testName, func(t *testing.T) {
			abort := revive(func() {
				Submit(cross, testInput.x, testInput.y, "R")
			})

			if abort == nil {
				t.Fatal("Invalid Submit didn't panic")
			}
		})
	}
}

func Test_submitInvalidColor(t *testing.T) {
	for _, testInput := range testSubmitInvalidColor {
		t.Run(testInput, func(t *testing.T) {
			abort := revive(func() {
				Submit(cross, 0, 0, testInput)
			})

			if abort == nil {
				t.Fatal("Invalid Submit didn't panic")
			}
		})
	}
}

func Test_submitValid(t *testing.T) {
	tableColor[39][1] = Color{"rgb(250,250,250)", "W"}

	Submit(cross, 1, 39, "R")

	if tableColor[39][1].initial != "R" {
		t.Fatal("Submit not taken into account")
	}
}

func Test_parseXYValid(t *testing.T) {
	for _, testInput := range testValidAdress {
		t.Run(testInput.address, func(t *testing.T) {
			x, y, err := parseXY(testInput.address)

			if err != nil || x != testInput.x || y != testInput.y {
				t.Fatal("coordinates should be valid")
			}
		})
	}
}

func Test_parseXYInvalid(t *testing.T) {
	for _, testInput := range testInvalidAdress {
		t.Run(testInput, func(t *testing.T) {
			_, _, err := parseXY(testInput)

			if err == nil {
				t.Fatal("coordinates should be invalid")
			}
		})
	}
}
