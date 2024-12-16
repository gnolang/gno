package main

import (
	"crypto/sha256"
	"fmt"
)

func sha256Crypt(password, salt, magic string) []byte {
	toHash := password + magic + salt
	mixin := sha256.Sum256([]byte(password + salt))

	for i := 0; i < len(password); i++ {
		toHash += string(mixin[i%32])
	}

	res := sha256.Sum256([]byte(toHash))
	return res[:]
}

func main() {
	b := sha256Crypt("1", "2", "3")

	fmt.Println(b)
}

// Output:
// [172 65 148 29 23 72 77 86 46 80 184 188 192 158 154 11 145 11 197 253 206 210 141 253 188 27 157 126 89 142 179 143]
