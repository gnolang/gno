package flippando

import (
	"std"
	"testing"
	"strings"
	"encoding/base64"

	"gno.land/p/demo/testutils"
)

func TestGenerateGrid(t *testing.T) {
    solvedBoard := generateSolvedBoard()

    tests := []struct {
        name         string
        solvedBoard  []int64
        gridType     string
        expectError  bool
    }{
        {"ValidSquareGrid", solvedBoard, "squareGrid", false},
        {"ValidDiceGrid", solvedBoard, "dice", false},
        {"ValidHexagramGrid", solvedBoard, "hexagrams", false},
        {"InvalidGridType", solvedBoard, "invalidType", true},
        {"InvalidBoardSize", make([]int64, 10), "squareGrid", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := GenerateGrid(tt.solvedBoard, tt.gridType)
            if (err != nil) != tt.expectError {
                t.Errorf("GenerateGrid() error = %v, expectError %v", err, tt.expectError)
            }
        })
    }
}

func TestGenerateColorSquareGrid(t *testing.T) {
    // Example solved board
    solvedBoard := generateSolvedBoard()

    // Call the function with the test board
    base64Svg, err := generateColorSquareGrid(solvedBoard, "squareGrid", 4) // gridType squareGrid, size 4 for a 4x4 grid
    if err != nil {
        t.Fatalf("generateColorSquareGrid returned an unexpected error: %v", err)
    }

    // Decode the Base64-encoded SVG string
    svgBytes, err := base64.StdEncoding.DecodeString(base64Svg)
    if err != nil {
        t.Fatalf("Error decoding Base64 SVG string: %v", err)
    }
    svg := string(svgBytes)

    // Check if the SVG string starts with <svg and ends with </svg>
    if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(svg, "</svg>") {
        t.Errorf("SVG does not start with '<svg' or end with '</svg>'")
    }

    // Count the number of <rect> elements
    rectCount := strings.Count(svg, "<rect")
    expectedRectCount := len(solvedBoard) // There should be a rect for each board element
    if rectCount != expectedRectCount {
        t.Errorf("Expected %d <rect> elements, found %d", expectedRectCount, rectCount)
    }
}


func generateSolvedBoard() []int64 {
    board := make([]int64, 16)
    for i := 0; i < 4; i++ {
        for j := 1; j <= 4; j++ {
            board[i*4+j-1] = int64(j)
        }
    }
    return board
}
