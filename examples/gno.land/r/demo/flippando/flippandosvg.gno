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
	case "squareGrid":
		return generateColorSquareGrid(solvedBoard, size)
	case "dice":
		return generateDiceGrid(solvedBoard, size)
	case "hexagrams":
		return generateHexagramGrid(solvedBoard, size)
	default:
		return "", errors.New("invalid grid type")
	}
}


func generateColorSquareGrid(solvedBoard []int64, size int) (string, error) {
	svgSquare := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">`
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
	svg.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">`)

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
	svg.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">`)

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

/*

function generateHexagramGrid(uint8[] memory solvedBoard, uint256 size) internal pure returns (string memory) {
    Hexagram memory hexagram = Hexagram({
        h1: "<svg viewBox="0 0 25 25" width="25" height="25"><line x1="4" y1="6" x2="10" y2="6" stroke="black" strokeWidth="2" /><line x1="15" y1="6" x2="21" y2="6" stroke="black" strokeWidth="2" /><line x1="4" y1="9" x2="21" y2="9" stroke="black" strokeWidth="2" /><line x1="4" y1="12" x2="10" y2="12" stroke="black" strokeWidth="2" /><line x1="15" y1="12" x2="21" y2="12" stroke="black" strokeWidth="2" /><line x1="4" y1="15" x2="21" y2="15" stroke="black" strokeWidth="2" /><line x1="4" y1="18" x2="10" y2="18" stroke="black" strokeWidth="2" /><line x1="15" y1="18" x2="21" y2="18" stroke="black" strokeWidth="2" /><line x1="4" y1="21" x2="21" y2="21" stroke="black" strokeWidth="2" /></svg>",
        h2: "<svg viewBox="0 0 25 25" width="25" height="25"><line x1="4" y1="6" x2="10" y2="6" stroke="black" strokeWidth="2" /><line x1="15" y1="6" x2="21" y2="6" stroke="black" strokeWidth="2" /><line x1="4" y1="9" x2="21" y2="9" stroke="black" strokeWidth="2" /><line x1="4" y1="12" x2="10" y2="12" stroke="black" strokeWidth="2" /><line x1="15" y1="12" x2="21" y2="12" stroke="black" strokeWidth="2" /><line x1="4" y1="15" x2="21" y2="15" stroke="black" strokeWidth="2" /><line x1="4" y1="18" x2="21" y2="18" stroke="black" strokeWidth="2" /><line x1="4" y1="21" x2="10" y2="21" stroke="black" strokeWidth="2" /><line x1="15" y1="21" x2="21" y2="21" stroke="black" strokeWidth="2" /></svg>",
        h3: "<svg viewBox="0 0 25 25" width="25" height="25"><line x1="4" y1="6" x2="10" y2="6" stroke="black" strokeWidth="2" /><line x1="15" y1="6" x2="21" y2="6" stroke="black" strokeWidth="2" /><line x1="4" y1="9" x2="21" y2="9" stroke="black" strokeWidth="2" /><line x1="4" y1="12" x2="10" y2="12" stroke="black" strokeWidth="2" /><line x1="15" y1="12" x2="21" y2="12" stroke="black" strokeWidth="2" /><line x1="4" y1="15" x2="21" y2="15" stroke="black" strokeWidth="2" /><line x1="4" y1="18" x2="21" y2="18" stroke="black" strokeWidth="2" /><line x1="4" y1="21" x2="21" y2="21" stroke="black" strokeWidth="2" /></svg>",
        h4: "<svg viewBox="0 0 25 25" width="25" height="25"><line x1="4" y1="6" x2="10" y2="6" stroke="black" strokeWidth="2" /><line x1="15" y1="6" x2="21" y2="6" stroke="black" strokeWidth="2" /><line x1="4" y1="9" x2="21" y2="9" stroke="black" strokeWidth="2" /><line x1="4" y1="12" x2="21" y2="12" stroke="black" strokeWidth="2" /><line x1="4" y1="15" x2="10" y2="15" stroke="black" strokeWidth="2" /><line x1="15" y1="15" x2="21" y2="15" stroke="black" strokeWidth="2" /><line x1="4" y1="18" x2="21" y2="18" stroke="black" strokeWidth="2" /><line x1="4" y1="21" x2="10" y2="21" stroke="black" strokeWidth="2" /><line x1="15" y1="21" x2="21" y2="21" stroke="black" strokeWidth="2" /></svg>"
    });

    string memory svg = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">';
    
    for (uint256 i = 0; i < size; i++) {
        for (uint256 j = 0; j < size; j++) {
            
            uint256 index = i * size + j;
            string memory hexagramType;

            if (solvedBoard[index] == 1){
                hexagramType = hexagram.h1;
            }
            else if (solvedBoard[index] == 2){
                hexagramType = hexagram.h2;
            }
            else if (solvedBoard[index] == 3){
                hexagramType = hexagram.h3;
            }
            else if (solvedBoard[index] == 4){
                hexagramType = hexagram.h4;
            }
            svg = string(abi.encodePacked(svg, '<g transform="translate(', uintToString(j * 25), ' ', uintToString(i * 25), ')">', hexagramType, '</g>'));
        }
    }

    svg = string(abi.encodePacked(svg, '</svg>'));
    bytes memory svgBytes = bytes(svg);
    string memory base64Svg = Base64.encode(svgBytes);
    return base64Svg;
}*/

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
