// Package faker provides fake data generation utilities for testing and
// development purposes. It includes methods for generating names, addresses,
// dates, emails, and other common data types using a pseudorandom generator.
package faker

import (
	"crypto/bech32"
	"math/rand"
	"strconv"
	"strings"
	"unicode"

	"gno.land/p/demo/entropy"
	"gno.land/p/nt/ufmt"
)

// Faker provides fake data generation using a pseudorandom number generator.
type Faker struct {
	rng *rand.Rand
}

// NewGenerator creates a new Faker instance with a random seed.
func NewGenerator() *Faker {
	return NewGeneratorWithSeed(entropy.New().Value64())
}

// NewGeneratorWithSeed creates a new Faker instance with a specified seed.
func NewGeneratorWithSeed(seed uint64) *Faker {
	return &Faker{
		rng: rand.New(rand.NewPCG(seed, 0xdeadbeef)),
	}
}

// FirstName returns a random first name.
func (f *Faker) FirstName() string {
	return f.pickRandom(&FirstNames)
}

// LastName returns a random last name.
func (f *Faker) LastName() string {
	return f.pickRandom(&LastNames)
}

// FullName returns a random full name (first name + last name).
func (f *Faker) FullName() string {
	return f.FirstName() + " " + f.LastName()
}

// Age returns a random age between min and min+80.
func (f *Faker) Age(min int) int {
	return f.rng.IntN(80) + min
}

// City returns a random city name.
func (f *Faker) City() string {
	return f.pickRandom(&Cities)
}

// Country returns a random country name.
func (f *Faker) Country() string {
	return f.pickRandom(&Countries)
}

// Address returns a random address in the format "123 Street Name, City, Country".
func (f *Faker) Address() string {
	var (
		streetNum  = strconv.Itoa(f.rng.IntN(9999) + 1)
		streetName = f.pickRandom(&StreetNames)
		city       = f.City()
		country    = f.Country()
	)

	return ufmt.Sprintf("%s %s, %s, %s", streetNum, streetName, city, country)
}

// Lorem returns random lorem ipsum text with the specified number of words.
func (f *Faker) Lorem(wordCount int) string {
	if wordCount <= 0 {
		return ""
	}

	result := ""

	for i := 0; i < wordCount; i++ {
		if i > 0 {
			result += " "
		}
		result += f.pickRandom(&LoremWords)
	}

	return result
}

// LoremSentence returns a random lorem ipsum sentence with 5-15 words.
func (f *Faker) LoremSentence() string {
	var (
		wordCount = f.rng.IntN(10) + 5
		sentence  = f.Lorem(wordCount)
	)

	if len(sentence) > 0 {
		sentence = string(sentence[0]-32) + sentence[1:] + "."
	}

	return sentence
}

// LoremParagraph returns a random lorem ipsum paragraph with 3-8 sentences.
func (f *Faker) LoremParagraph() string {
	var (
		sentenceCount = f.rng.IntN(5) + 3
		paragraph     = ""
	)

	for i := 0; i < sentenceCount; i++ {
		if i > 0 {
			paragraph += " "
		}
		paragraph += f.LoremSentence()
	}

	return paragraph
}

// Date returns a random date in YYYY-MM-DD format between 1974 and 2025.
func (f *Faker) Date() string {
	var (
		year  = strconv.Itoa(f.rng.IntN(52) + 1974)
		month = itoa2Digits(f.rng.IntN(12) + 1)
		day   = itoa2Digits(f.rng.IntN(28) + 1) // Don't bother handling month lengths
	)

	return ufmt.Sprintf("%s-%s-%s", year, month, day)
}

// Time returns a random time in HH:MM:SS format (24-hour).
func (f *Faker) Time() string {
	var (
		hour   = itoa2Digits(f.rng.IntN(24))
		minute = itoa2Digits(f.rng.IntN(60))
		second = itoa2Digits(f.rng.IntN(60))
	)

	return ufmt.Sprintf("%s:%s:%s", hour, minute, second)
}

// DateTime returns a random date and time in "YYYY-MM-DD HH:MM:SS" format.
func (f *Faker) DateTime() string {
	return f.Date() + " " + f.Time()
}

// Email returns a random email address in firstname.lastname@domain format.
func (f *Faker) Email() string {
	var (
		firstName = strings.ToLower(f.FirstName())
		lastName  = strings.ToLower(f.LastName())
		domain    = f.pickRandom(&EmailDomains)
	)

	return ufmt.Sprintf("%s.%s@%s", firstName, lastName, domain)
}

// Phone returns a random phone number in (XXX) XXX-XXXX format.
func (f *Faker) Phone() string {
	var (
		area     = f.rng.IntN(900) + 100
		exchange = f.rng.IntN(900) + 100
		number   = f.rng.IntN(10000)
	)

	return ufmt.Sprintf("(%d) %d-%d", area, exchange, number)
}

// StdAddress returns a random chain.Address.
func (f *Faker) StdAddress() address {
	rawAddr := make([]byte, 32)
	for i := 0; i < 32; i++ {
		rawAddr[i] = byte(f.rng.IntN(32))
	}

	addr, err := bech32.Encode("g", rawAddr)
	if err != nil {
		panic(err)
	}

	return address(addr)
}

// Username returns a random gno username following the rules:
// - Start with minimum 3 letters (lowercase)
// - End with minimum 3 numbers
// - Less than 20 chars total
// - Only letters, numbers, and underscore allowed
func (f *Faker) Username() string {
	var (
		username = ""
		sources  = []*[]string{
			&FirstNames,
			&LastNames,
			&LoremWords,
		}
	)

	for {
		var (
			source = sources[f.rng.IntN(len(sources))]
			word   = strings.ToLower(f.pickRandomLettersOnly(source))
			next   = username + word
		)

		if len(next) > 16 && len(username) > 3 {
			break
		} else if len(next) <= 16 {
			username = next
		}
	}

	for i := 0; i < 3; i++ {
		username += strconv.Itoa(f.rng.IntN(10))
	}

	return username
}

func (f *Faker) pickRandom(source *[]string) string {
	return (*source)[f.rng.IntN(len(*source))]
}

func (f *Faker) pickRandomLettersOnly(source *[]string) string {
	isLettersOnly := func(s string) bool {
		for _, r := range s {
			if !unicode.IsLetter(r) {
				return false
			}
		}
		return true
	}

	// Limit attempts to 100 to avoid infinite loops.
	for i := 0; i < 100; i++ {
		word := f.pickRandom(source)
		if isLettersOnly(word) {
			return word
		}
	}

	panic("could not find a letters-only word after 100 attempts")
}

func itoa2Digits(n int) string {
	str := strconv.Itoa(n)
	if n < 10 {
		return "0" + str
	}
	return str
}

// defaultFaker is the default Faker instance used by package-level functions.
var defaultFaker = NewGenerator()

// FirstName returns a random first name using the default faker.
func FirstName() string {
	return defaultFaker.FirstName()
}

// LastName returns a random last name using the default faker.
func LastName() string {
	return defaultFaker.LastName()
}

// FullName returns a random full name using the default faker.
func FullName() string {
	return defaultFaker.FullName()
}

// Age returns a random age using the default faker.
func Age(min int) int {
	return defaultFaker.Age(min)
}

// City returns a random city name using the default faker.
func City() string {
	return defaultFaker.City()
}

// Country returns a random country name using the default faker.
func Country() string {
	return defaultFaker.Country()
}

// Address returns a random address using the default faker.
func Address() string {
	return defaultFaker.Address()
}

// Lorem returns random lorem ipsum text using the default faker.
func Lorem(wordCount int) string {
	return defaultFaker.Lorem(wordCount)
}

// LoremSentence returns a random lorem ipsum sentence using the default faker.
func LoremSentence() string {
	return defaultFaker.LoremSentence()
}

// LoremParagraph returns a random lorem ipsum paragraph using the default faker.
func LoremParagraph() string {
	return defaultFaker.LoremParagraph()
}

// Date returns a random date using the default faker.
func Date() string {
	return defaultFaker.Date()
}

// Time returns a random time using the default faker.
func Time() string {
	return defaultFaker.Time()
}

// DateTime returns a random date and time using the default faker.
func DateTime() string {
	return defaultFaker.DateTime()
}

// Email returns a random email address using the default faker.
func Email() string {
	return defaultFaker.Email()
}

// Phone returns a random phone number using the default faker.
func Phone() string {
	return defaultFaker.Phone()
}

// StdAddress returns a random chain.Address using the default faker.
func StdAddress() address {
	return defaultFaker.StdAddress()
}

// Username returns a random gno username using the default faker.
func Username() string {
	return defaultFaker.Username()
}
