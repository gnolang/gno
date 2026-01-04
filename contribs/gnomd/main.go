package main

import (
	"fmt"
	"io"
	"os"

	markdown "github.com/MichaelMure/go-term-markdown"
)

func main() {
	// If no arguments are provided, read from stdin
	if len(os.Args) <= 1 {
		fileContent, err := io.ReadAll(os.Stdin)
		checkErr(err)
		renderMarkdown("stdin.gno", fileContent)
	}

	// Iterate through command-line arguments (file paths)
	for _, filePath := range os.Args[1:] {
		fileContent, err := os.ReadFile(filePath)
		checkErr(err)
		renderMarkdown(filePath, fileContent)
	}
}

func renderMarkdown(filePath string, fileContent []byte) {
	fmt.Printf("-- %s --\n", filePath)

	result := markdown.Render(string(fileContent), 80, 6)
	fmt.Println(string(result))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
