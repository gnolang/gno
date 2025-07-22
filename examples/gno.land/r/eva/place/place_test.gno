package place

import (
	"encoding/base64"
	"testing"
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
	_, err := generateSquareSvg("rgb(255555,0,0)")

	if err == nil {
		t.Fatal("color should be invalid")
	}
}

func Test_checkColorValid_true1(t *testing.T) {
	value := checkColorValid("rgb(255,0,0)")

	// Verify the svg is correct
	if value == false {
		t.Fatalf("The color should be valid")
	}
}

func Test_checkColorValid_true2(t *testing.T) {
	value := checkColorValid("rgb(0,0,0)")

	// Verify the svg is correct
	if value == false {
		t.Fatalf("The color should be valid")
	}
}

func Test_checkColorValid_true3(t *testing.T) {
	value := checkColorValid("rgb(255,255,255)")

	// Verify the svg is correct
	if value == false {
		t.Fatalf("The color should be valid")
	}
}

func Test_checkColorValid_false1(t *testing.T) {
	value := checkColorValid("rgb(,0,0)")

	// Verify the svg is correct
	if value == true {
		t.Fatalf("The color should be invalid")
	}
}

func Test_checkColorValid_false2(t *testing.T) {
	value := checkColorValid("rgb(1,0;0)")

	// Verify the svg is correct
	if value == true {
		t.Fatalf("The color should be invalid")
	}
}

func Test_checkColorValid_false3(t *testing.T) {
	value := checkColorValid("rgb(1k,0,0)")

	// Verify the svg is correct
	if value == true {
		t.Fatalf("The color should be invalid")
	}
}

func Test_checkColorValid_false4(t *testing.T) {
	value := checkColorValid("rgb(1222,0,0)")

	// Verify the svg is correct
	if value == true {
		t.Fatalf("The color should be invalid")
	}
}

func Test_submitInvalidCoordinates1(t *testing.T) {
	abort := revive(func() {
		Submit(cross, 1000, 1, "R")
	})

	if abort == nil {
		t.Fatal("Invalid Submit didn't panic")
	}
}

func Test_submitInvalidCoordinates2(t *testing.T) {
	abort := revive(func() {
		Submit(cross, 1, 40, "R") // 40 is the number of colomns / rows, the last assignable value is 39
	})

	if abort == nil {
		t.Fatal("Invalid Submit didn't panic")
	}
}

func Test_submitInvalidColor(t *testing.T) {
	abort := revive(func() {
		Submit(cross, 1, 40, "This color is definitly not a valid one")
	})

	if abort == nil {
		t.Fatal("Invalid Submit didn't panic")
	}
}

func Test_submitInvalidColorEmpty(t *testing.T) {
	abort := revive(func() {
		Submit(cross, 1, 40, "")
	})

	if abort == nil {
		t.Fatal("Invalid Submit didn't panic")
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
	x, y, err := parseXY("someAdress/r/eva/place:?color=G&x=9&y=2")

	if err != nil || x != 9 || y != 2 {
		t.Fatal("coordinates should be valid")
	}
}

func Test_parseXYValid2(t *testing.T) {
	x, y, err := parseXY("?x=19&y=1")

	if err != nil || x != 19 || y != 1 {
		t.Fatal("coordinates should be valid")
	}
}

func Test_parseXYInvalid(t *testing.T) {
	_, _, err := parseXY("someAdress/r/eva/place:?color=G&x=9a&y=2")

	if err == nil {
		t.Fatal("coordinates should be invalid")
	}
}

func Test_parseXYInvalid2(t *testing.T) {
	_, _, err := parseXY("")

	if err == nil {
		t.Fatal("coordinates should be invalid")
	}
}
