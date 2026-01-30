package place

import (
	"errors"
	"strconv"
	"strings"

	"gno.land/p/demo/svg"
)

const (
	minColorLength = 10 // exemple rgb(0,0,0)
	maxColorLength = 16 // exemple rgb(255,255,255)
)

func checkColorValid(color string) bool {
	if len(color) < minColorLength || len(color) > maxColorLength ||
		color[:4] != "rgb(" || color[len(color)-1] != ')' {
		return false
	}

	numbers := strings.Split(color[4:len(color)-1], ",")
	if len(numbers) != 3 {
		return false
	}

	for _, numberString := range numbers {
		numberValue, err := strconv.ParseUint(numberString, 10, 0)
		if err != nil || numberValue > 255 {
			return false
		}
	}
	return true
}

func findColor(colorName string) (bool, Color) {
	for _, color := range colors {
		if color.initial == colorName {
			return true, color
		}
	}
	return false, Color{}
}

func generateSquareSvg(color string) (string, error) {
	if checkColorValid(color) == false {
		return "", errors.New("Invalid color")
	}
	canvas := svg.Canvas{Width: 30, Height: 30}
	canvas.Append(svg.NewRectangle(0, 0, 30, 30, color))
	return canvas.Base64(), nil
}

func generateSVG() string {
	canvas := svg.Canvas{Width: (int)(cols * 40), Height: (int)(rows * 40)}
	for y := 0; y < int(rows); y++ {
		for x := 0; x < int(cols); x++ {
			canvas.Append(svg.NewRectangle(x*40, y*40, 40, 40, tableColor[y][x].rgb))
		}
	}
	return canvas.Base64()
}
