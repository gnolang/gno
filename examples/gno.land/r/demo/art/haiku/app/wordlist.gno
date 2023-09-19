package haiku

import (
	"gno.land/r/demo/art/haiku/wordlist0"
	"gno.land/r/demo/art/haiku/wordlist1"
	"gno.land/r/demo/art/haiku/wordlist2"
	"gno.land/r/demo/art/haiku/wordlist3"
)

func syllablesInWord(s string) (int, bool) {
	fns := []func(string) (int, bool){
		wordlist0.SyllablesInWord, wordlist1.SyllablesInWord, wordlist2.SyllablesInWord, wordlist3.SyllablesInWord,
	}
	for _, fn := range fns {
		n, ok := fn(s)
		if ok {
			return n, ok
		}
	}
	return 0, false
}
