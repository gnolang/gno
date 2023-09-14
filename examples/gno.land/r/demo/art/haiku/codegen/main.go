package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	// the word list with syllables is obtained from
	// https://github.com/gautesolheim/25000-syllabified-words-list
	// which is released into the public domain.
	f, _ := os.Create("../wordlist.gno")
	f.WriteString(`package haiku
	
func syllablesInWord(s string) (int, bool) {
	switch s {
`)

	file, err := os.Open("wordlist.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := scanner.Text()
		s := strings.Replace(line, ";", "", -1)
		f.WriteString(fmt.Sprintf("\t\tcase \"%s\":\n\t\t\treturn %d, true\n", s, 1+strings.Count(line, ";")))
	}
	f.WriteString(`
	}
	return 0, false
}`)
	f.Close()

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
