package flippando

import (
	"strings"
	"strconv"
	"errors"
	"encoding/base64"

	"gno.land/p/demo/ufmt"
)

type Dice struct {
	d1, d2, d3, d4, d5, d6 string
}

type Hexagram struct {
	h1, h2, h3, h4 string
}

func GenerateGrid(solvedBoard []int64, gridType string) (string, error) {
	var size int
	switch len(solvedBoard) {
	case 16:
		size = 4
	case 64:
		size = 8
	case 256:
		size = 16
	default:
		return "", errors.New("invalid board size")
	}

	switch gridType {
	case "squareGrid", "greyGradient", "redGradient", "greenGradient", "blueGradient":
		return generateColorSquareGrid(solvedBoard, gridType, size)
	case "dice":
		return generateDiceGrid(solvedBoard, size)
	case "hexagrams":
		return generateHexagramGrid(solvedBoard, size)
	default:
		return "", errors.New("invalid grid type")
	}
}


func generateColorSquareGrid(solvedBoard []int64, gridType string, size int) (string, error) {
	svgSquare := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">`
	if size == 8 {
		svgSquare = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">`
	}
	
	for i := int64(0); i < int64(size); i++ {
		for j := int64(0); j < int64(size); j++ {
			index := i*int64(size) + j
			key := solvedBoard[index]
			color, _ := getColorFromKey(key, gridType)
			rect := ufmt.Sprintf(`<rect width="25" height="25" fill="%s" x="%s" y="%s" />`, color, intToString(j*25), intToString(i*25))
			svgSquare += rect
		}
	}
	svgSquare += `</svg>`
	base64Svg := base64.StdEncoding.EncodeToString([]byte(svgSquare))
	return base64Svg, nil
}

func generateDiceGrid(solvedBoard []int64, size int) (string, error) {
	dice := Dice{
		d1: `<svg width="25" height="25"><circle cx="12.5" cy="12.5" r="2.5" fill="black"/></svg>`,
		d2: `<svg width="25" height="25"><circle cx="6.25" cy="6.25" r="2.5" fill="black"/><circle cx="18.75" cy="18.75" r="2.5" fill="black"/></svg>`,
        d3: `<svg width="25" height="25"><circle cx="6.25" cy="6.25" r="2.5" fill="black"/><circle cx="12.5" cy="12.5" r="2.5" fill="black"/><circle cx="18.75" cy="18.75" r="2.5" fill="black"/></svg>`,
        d4: `<svg width="25" height="25"><circle cx="6.25" cy="6.25" r="2.5" fill="black"/><circle cx="18.75" cy="6.25" r="2.5" fill="black"/><circle cx="6.25" cy="18.75" r="2.5" fill="black"/><circle cx="18.75" cy="18.75" r="2.5" fill="black"/></svg>`,
        d5: `<svg width="25" height="25"><circle cx="6.25" cy="6.25" r="2.5" fill="black"/><circle cx="18.75" cy="6.25" r="2.5" fill="black"/><circle cx="12.5" cy="12.5" r="2.5" fill="black"/><circle cx="6.25" cy="18.75" r="2.5" fill="black"/><circle cx="18.75" cy="18.75" r="2.5" fill="black"/></svg>`,
        d6: `<svg width="25" height="25"><circle cx="6.25" cy="6.25" r="2.5" fill="black"/><circle cx="18.75" cy="6.25" r="2.5" fill="black"/><circle cx="6.25" cy="12.5" r="2.5" fill="black"/><circle cx="18.75" cy="12.5" r="2.5" fill="black"/><circle cx="6.25" cy="18.75" r="2.5" fill="black"/><circle cx="18.75" cy="18.75" r="2.5" fill="black"/></svg>`,
		
	}

	var svg strings.Builder
	svgSquare := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">`
	if size == 8 {
		svgSquare = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">`
	}
	svg.WriteString(svgSquare)

	for i := int(0); i < size; i++ {
		for j := int(0); j < size; j++ {
			index := i*size + j
			var diceFace string

			switch solvedBoard[index] {
			case 1:
				diceFace = dice.d1
			case 2:
				diceFace = dice.d2
            case 3:
				diceFace = dice.d3
			case 4:
				diceFace = dice.d4  
            case 5:
				diceFace = dice.d5
			case 6:
				diceFace = dice.d6  
			}

			svg.WriteString(ufmt.Sprintf(`<g transform="translate(%d %d)">%s</g>`, j*25, i*25, diceFace))
		}
	}

	svg.WriteString(`</svg>`)
	svgBytes := svg.String()
	base64Svg := base64.StdEncoding.EncodeToString([]byte(svgBytes))
	return base64Svg, nil
}

func generateHexagramGrid(solvedBoard []int64, size int) (string, error) {
	hexagram := Hexagram{
        h1: `<svg viewBox="0 0 25 25" width="25" height="25"><line x1="4" y1="6" x2="10" y2="6" stroke="black" strokeWidth="2" /><line x1="15" y1="6" x2="21" y2="6" stroke="black" strokeWidth="2" /><line x1="4" y1="9" x2="21" y2="9" stroke="black" strokeWidth="2" /><line x1="4" y1="12" x2="10" y2="12" stroke="black" strokeWidth="2" /><line x1="15" y1="12" x2="21" y2="12" stroke="black" strokeWidth="2" /><line x1="4" y1="15" x2="21" y2="15" stroke="black" strokeWidth="2" /><line x1="4" y1="18" x2="10" y2="18" stroke="black" strokeWidth="2" /><line x1="15" y1="18" x2="21" y2="18" stroke="black" strokeWidth="2" /><line x1="4" y1="21" x2="21" y2="21" stroke="black" strokeWidth="2" /></svg>`,
        h2: `<svg viewBox="0 0 25 25" width="25" height="25"><line x1="4" y1="6" x2="10" y2="6" stroke="black" strokeWidth="2" /><line x1="15" y1="6" x2="21" y2="6" stroke="black" strokeWidth="2" /><line x1="4" y1="9" x2="21" y2="9" stroke="black" strokeWidth="2" /><line x1="4" y1="12" x2="10" y2="12" stroke="black" strokeWidth="2" /><line x1="15" y1="12" x2="21" y2="12" stroke="black" strokeWidth="2" /><line x1="4" y1="15" x2="21" y2="15" stroke="black" strokeWidth="2" /><line x1="4" y1="18" x2="21" y2="18" stroke="black" strokeWidth="2" /><line x1="4" y1="21" x2="10" y2="21" stroke="black" strokeWidth="2" /><line x1="15" y1="21" x2="21" y2="21" stroke="black" strokeWidth="2" /></svg>`,
        h3: `<svg viewBox="0 0 25 25" width="25" height="25"><line x1="4" y1="6" x2="10" y2="6" stroke="black" strokeWidth="2" /><line x1="15" y1="6" x2="21" y2="6" stroke="black" strokeWidth="2" /><line x1="4" y1="9" x2="21" y2="9" stroke="black" strokeWidth="2" /><line x1="4" y1="12" x2="10" y2="12" stroke="black" strokeWidth="2" /><line x1="15" y1="12" x2="21" y2="12" stroke="black" strokeWidth="2" /><line x1="4" y1="15" x2="21" y2="15" stroke="black" strokeWidth="2" /><line x1="4" y1="18" x2="21" y2="18" stroke="black" strokeWidth="2" /><line x1="4" y1="21" x2="21" y2="21" stroke="black" strokeWidth="2" /></svg>`,
        h4: `<svg viewBox="0 0 25 25" width="25" height="25"><line x1="4" y1="6" x2="10" y2="6" stroke="black" strokeWidth="2" /><line x1="15" y1="6" x2="21" y2="6" stroke="black" strokeWidth="2" /><line x1="4" y1="9" x2="21" y2="9" stroke="black" strokeWidth="2" /><line x1="4" y1="12" x2="21" y2="12" stroke="black" strokeWidth="2" /><line x1="4" y1="15" x2="10" y2="15" stroke="black" strokeWidth="2" /><line x1="15" y1="15" x2="21" y2="15" stroke="black" strokeWidth="2" /><line x1="4" y1="18" x2="21" y2="18" stroke="black" strokeWidth="2" /><line x1="4" y1="21" x2="10" y2="21" stroke="black" strokeWidth="2" /><line x1="15" y1="21" x2="21" y2="21" stroke="black" strokeWidth="2" /></svg>`,
	}

	var svg strings.Builder
	svgSquare := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">`
	if size == 8 {
		svgSquare = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">`
	}
	svg.WriteString(svgSquare)

	for i := int(0); i < size; i++ {
		for j := int(0); j < size; j++ {
			index := i*size + j
			var hexagramType string

			switch solvedBoard[index] {
			case 1:
				hexagramType = hexagram.h1
			case 2:
				hexagramType = hexagram.h2
			case 3:
				hexagramType = hexagram.h3
			case 4:
				hexagramType = hexagram.h4
			}

			svg.WriteString(ufmt.Sprintf(`<g transform="translate(%d %d)">%s</g>`, j*25, i*25, hexagramType))
		}
	}

	svg.WriteString(`</svg>`)
    svgBytes := svg.String()
	base64Svg := base64.StdEncoding.EncodeToString([]byte(svgBytes))
	return base64Svg, nil

}

// util
func intToString(value int64) string {
	formatted := strconv.FormatUint(uint64(value), 10)
	return formatted;
}

func getColorFromKey(key int64, gridType string) (string, error) {
    switch key {
    case int64(1):
		switch gridType {
		case "squareGrid":
			return "#D9362A", nil
		case "greyGradient":
			return "#101010", nil
		case "redGradient":
			return "#800000", nil
		case "greenGradient":
			return "#008000", nil
		case "blueGradient":
			return "#000080", nil
		}
    case int64(2):
		switch gridType {
		case "squareGrid":
			return "#BE185D", nil
		case "greyGradient":
			return "#303030", nil
		case "redGradient":
			return "#901010", nil
		case "greenGradient":
			return "#109010", nil
		case "blueGradient":
			return "#101090", nil
		}
        
    case int64(3):
		switch gridType {
		case "squareGrid":
			return "#EC1818", nil
		case "greyGradient":
			return "#505050", nil
		case "redGradient":
			return "#a02020", nil
		case "greenGradient":
			return "#20a020", nil
		case "blueGradient":
			return "#2020a0", nil
		}
        
    case int64(4):
		switch gridType {
		case "squareGrid":
			return "#F4AA24", nil
		case "greyGradient":
			return "#707070", nil
		case "redGradient":
			return "#b03030", nil
		case "greenGradient":
			return "#30b030", nil
		case "blueGradient":
			return "#3030b0", nil
		}
        
    case int64(5):
		switch gridType {
		case "squareGrid":
			return "#F4D424", nil
		case "greyGradient":
			return "#909090", nil
		case "redGradient":
			return "#c04040", nil
		case "greenGradient":
			return "#40c040", nil
		case "blueGradient":
			return "#4040c0", nil
		}
        
    case int64(6):
		switch gridType {
		case "squareGrid":
			return "#F1F223", nil
		case "greyGradient":
			return "#b0b0b0", nil
		case "redGradient":
			return "#d05050", nil
		case "greenGradient":
			return "#50d050", nil
		case "blueGradient":
			return "#5050d0", nil
		}
        
    case int64(7):
		switch gridType {
		case "squareGrid":
			return "#A5DD0C", nil
		case "greyGradient":
			return "#d0d0d0", nil
		case "redGradient":
			return "#e06060", nil
		case "greenGradient":
			return "#60e060", nil
		case "blueGradient":
			return "#6060e0", nil
		}
        
    case int64(8):
		switch gridType {
		case "squareGrid":
			return "#37B400", nil
		case "greyGradient":
			return "#f0f0f0", nil
		case "redGradient":
			return "#f07070", nil
		case "greenGradient":
			return "#70f070", nil
		case "blueGradient":
			return "#7070f0", nil
		}
        
    case int64(9):
		switch gridType {
		case "squareGrid":
        	return "#98D0E9", nil
		}

    case int64(10):
		switch gridType {
		case "squareGrid":
        	return "#3131FD", nil
		}
        
    case int64(11):
		switch gridType {
		case "squareGrid":
        	return "#1D2C85", nil
		}
        
    case int64(12):
		switch gridType {
		case "squareGrid":
        	return "#F2F2F2", nil
		}
        
    case int64(13):
		switch gridType {
		case "squareGrid":
        	return "#1A1A1A", nil
		}
        
    case int64(14):
		switch gridType {
		case "squareGrid":
        	return "#71842F", nil
		}
        
    case int64(15):
		switch gridType {
		case "squareGrid":
        	return "#90572F", nil
		}
        
    case int64(16):
		switch gridType {
		case "squareGrid":
        	return "#8316C0", nil
		}
        
    default:
        panic("Invalid color key " + intToString(key) + " " + gridType)
    }
	return "", nil
}
