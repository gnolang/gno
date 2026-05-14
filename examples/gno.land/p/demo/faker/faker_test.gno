package faker

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
)

func TestNewGenerator(t *testing.T) {
	f := NewGenerator()

	uassert.NotNil(t, f)
	uassert.NotNil(t, f.rng)
}

func TestNewGeneratorWithSeed(t *testing.T) {
	var (
		seed = uint64(12345)
		f    = NewGeneratorWithSeed(seed)
	)

	uassert.NotNil(t, f)
	uassert.NotNil(t, f.rng)
}

func TestDeterministicOutput(t *testing.T) {
	var (
		seed = uint64(42)
		f1   = NewGeneratorWithSeed(seed)
		f2   = NewGeneratorWithSeed(seed)
	)

	for i := 0; i < 42; i++ {
		uassert.Equal(t, f1.FirstName(), f2.FirstName())
		uassert.Equal(t, f1.LastName(), f2.LastName())
		uassert.Equal(t, f1.FullName(), f2.FullName())
		uassert.Equal(t, f1.Age(0), f2.Age(0))
		uassert.Equal(t, f1.City(), f2.City())
		uassert.Equal(t, f1.Country(), f2.Country())
		uassert.Equal(t, f1.Address(), f2.Address())
		uassert.Equal(t, f1.Lorem(42), f2.Lorem(42))
		uassert.Equal(t, f1.LoremSentence(), f2.LoremSentence())
		uassert.Equal(t, f1.LoremParagraph(), f2.LoremParagraph())
		uassert.Equal(t, f1.Date(), f2.Date())
		uassert.Equal(t, f1.Time(), f2.Time())
		uassert.Equal(t, f1.DateTime(), f2.DateTime())
		uassert.Equal(t, f1.Email(), f2.Email())
		uassert.Equal(t, f1.Phone(), f2.Phone())
		uassert.Equal(t, f1.StdAddress(), f2.StdAddress())
		uassert.Equal(t, f1.Username(), f2.Username())
	}
}

func TestPickRandomLettersOnly(t *testing.T) {
	f := NewGenerator()

	// Source without any letters only word
	testData := []string{"hello8", "123", "6world", "hey_hey", "ok."}

	uassert.PanicsWithMessage(
		t,
		"could not find a letters-only word after 100 attempts",
		func() { f.pickRandomLettersOnly(&testData) },
	)
}

type goldenTestCase struct {
	seed       uint64
	iterations []goldenValues
}

type goldenValues struct {
	firstName      string
	lastName       string
	fullName       string
	age            int
	city           string
	country        string
	address        string
	lorem          string
	loremSentence  string
	loremParagraph string
	date           string
	time           string
	dateTime       string
	email          string
	phone          string
	stdAddress     address
	username       string
}

func TestGoldenValues(t *testing.T) {
	for _, testCase := range goldenTestCases { // See golden_test_cases.gno
		f := NewGeneratorWithSeed(testCase.seed)

		for i, c := range testCase.iterations {
			uassert.Equal(
				t,
				c.firstName,
				f.FirstName(),
				ufmt.Sprintf("First name doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.lastName,
				f.LastName(),
				ufmt.Sprintf("Last name doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.fullName,
				f.FullName(),
				ufmt.Sprintf("Full name doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.age,
				f.Age(0),
				ufmt.Sprintf("Age doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.city,
				f.City(),
				ufmt.Sprintf("City doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.country,
				f.Country(),
				ufmt.Sprintf("Country doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.address,
				f.Address(),
				ufmt.Sprintf("Address doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.lorem,
				f.Lorem(42),
				ufmt.Sprintf("Lorem doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.loremSentence,
				f.LoremSentence(),
				ufmt.Sprintf("Lorem sentence doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.loremParagraph,
				f.LoremParagraph(),
				ufmt.Sprintf("Lorem paragraph doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.date,
				f.Date(),
				ufmt.Sprintf("Date doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.time,
				f.Time(),
				ufmt.Sprintf("Time doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.dateTime,
				f.DateTime(),
				ufmt.Sprintf("DateTime doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.email,
				f.Email(),
				ufmt.Sprintf("Email doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.phone,
				f.Phone(),
				ufmt.Sprintf("Phone doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.stdAddress,
				f.StdAddress(),
				ufmt.Sprintf("StdAddress doesn't match for seed %d (%d)", testCase.seed, i),
			)

			uassert.Equal(
				t,
				c.username,
				f.Username(),
				ufmt.Sprintf("Username doesn't match for seed %d (%d)", testCase.seed, i),
			)
		}
	}
}
