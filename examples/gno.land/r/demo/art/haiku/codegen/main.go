package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	lineCount := linesInFile("wordlist.txt")
	numSplits := 4
	for i := 0; i < numSplits; i++ {
		err := createSplit(i, lineCount, numSplits)
		if err != nil {
			log.Fatal(err)
		}
	}
	createMain(numSplits)
}

func linesInFile(fname string) int {
	file, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	count := 0
	for scanner.Scan() {
		count++
	}
	file.Close()
	return count
}

func createMain(numSplits int) (err error) {
	f, err := os.Create("../app/wordlist.gno")
	if err != nil {
		return
	}
	f.WriteString(`package haiku

	import (
`)
	for i := 0; i < numSplits; i++ {
		f.WriteString(fmt.Sprintf("\t\"gno.land/r/demo/art/haiku/wordlist%d\"\n", i))
	}
	f.WriteString(`)

	func syllablesInWord(s string) (int, bool) {
		fns := []func(string) (int, bool){
`)
	for i := 0; i < numSplits; i++ {
		f.WriteString(fmt.Sprintf("wordlist%d.SyllablesInWord", i))
		if i < numSplits-1 {
			f.WriteString(",")
		}
	}
	f.WriteString(`}
		for _, fn := range fns {
			n, ok := fn(s)
			if ok {
				return n, ok
			}
		}
		return 0, false
	}
	`)
	f.Close()
	return
}

func createSplit(i int, lineCount int, numSplits int) (err error) {
	realmPath := fmt.Sprintf("../wordlist%d/", i)
	err = os.MkdirAll(realmPath, os.ModePerm)
	if err != nil {
		return
	}

	f2, err := os.Create(fmt.Sprintf("%sgno.mod", realmPath))
	if err != nil {
		return
	}
	f2.WriteString(fmt.Sprintf("module gno.land/r/demo/art/haiku/wordlist%d\n", i))
	f2.Close()

	f, err := os.Create(fmt.Sprintf("%swordlist.gno", realmPath))
	if err != nil {
		return
	}
	f.WriteString(fmt.Sprintf("package wordlist%d\n\n", i))
	f.WriteString(`func SyllablesInWord(s string) (int, bool) {
	switch s {
`)

	// the word list with syllables is obtained from
	// https://github.com/gautesolheim/25000-syllabified-words-list
	// which is released into the public domain.
	file, err := os.Open("wordlist.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
		if count > i*lineCount/numSplits && count <= (i+1)*lineCount/numSplits {
			line := scanner.Text()
			s := strings.Replace(line, ";", "", -1)
			f.WriteString(fmt.Sprintf("\t\tcase \"%s\":\n\t\t\treturn %d, true\n", s, 1+strings.Count(line, ";")))
		}
	}
	f.WriteString(`
	}
	return 0, false
}`)
	f.Close()

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return
}
