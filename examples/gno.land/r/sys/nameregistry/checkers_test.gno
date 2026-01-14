package nameregistry_test

import (
	"testing"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/uassert"

	"gno.land/r/sys/nameregistry"
)

func TestNewChecker(t *testing.T) {
	uassert.PanicsWithMessage(t, "checker description is required", func() {
		_ = nameregistry.NewChecker("", nil)
	}, "expect required panic")

	uassert.PanicsWithMessage(t, "checker function is required", func() {
		_ = nameregistry.NewChecker("description", nil)
	}, "expect required panic")

	checker := nameregistry.NewChecker("Foobar", func(string) error { return nil })
	uassert.Equal(t, "Foobar", checker.Description(), "expect description to match")
}

func TestNewCheckerCheck(t *testing.T) {
	tests := []struct {
		name         string
		registerName string
		err          error
	}{
		{
			name:         "ok",
			registerName: "xyz123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var validatedName string
			checker := nameregistry.NewChecker("Foobar", func(name string) error {
				validatedName = name
				return tt.err
			})

			// Act
			err := checker.Check(tt.registerName)

			// Assert
			if tt.err != nil {
				uassert.ErrorIs(t, err, tt.err, "expect an error")
			} else {
				uassert.Nil(t, err, "expect no error")
			}

			uassert.Equal(t, tt.registerName, validatedName, "expect validated name to match")
		})
	}
}

func TestCheckers(t *testing.T) {
	// Min length
	uassert.PanicsWithMessage(t, "min length checker requires a length > 0", func() {
		_ = nameregistry.NewMinLengthChecker(0)
	}, "expect invalid length panic")
	checker := nameregistry.NewMinLengthChecker(4)
	uassert.Equal(t, "Minimum name length allowed is 4", checker.Description(), "expect description to match")

	// Max length
	uassert.PanicsWithMessage(t, "max length checker requires a length > 1", func() {
		_ = nameregistry.NewMaxLengthChecker(0)
	}, "expect invalid length panic")
	checker = nameregistry.NewMaxLengthChecker(4)
	uassert.Equal(t, "Maximum name length allowed is 4", checker.Description(), "expect description to match")

	// Lowercase
	checker = nameregistry.NewLowercaseChecker()
	uassert.Equal(t, "Alphanumeric chars must all have lowercase chars", checker.Description(), "expect description to match")

	// Forbidden prefix
	uassert.PanicsWithMessage(t, "name prefix checker requires at least one prefix", func() {
		_ = nameregistry.NewForbiddenPrefixChecker()
	}, "expect panic when no prefix")
	checker = nameregistry.NewForbiddenPrefixChecker("a", "b")
	uassert.Equal(t, "Name must not start with: a, b", checker.Description(), "expect description to match")

	// Alphanum
	checker = nameregistry.NewAlphaNumChecker()
	uassert.Equal(t, "Name must start with a letter and end with 3 numbers", checker.Description(), "expect description to match")

	// Homoglyph
	uassert.PanicsWithMessage(t, "homoglyph checker requires at least one names tree", func() {
		_ = nameregistry.NewHomoglyphChecker()
	}, "expect panic when no name tree")
	checker = nameregistry.NewHomoglyphChecker(avl.NewTree())
	uassert.Equal(t, "Reserved or blacklisted names are not allowed", checker.Description(), "expect description to match")
}

func TestCheckersCheck(t *testing.T) {
	tests := []struct {
		name         string
		registerName string
		setup        func() nameregistry.Checker
		errMsg       string
	}{
		{
			name: "min length valid",
			setup: func() nameregistry.Checker {
				return nameregistry.NewMinLengthChecker(4)
			},
			registerName: "x123",
		},
		{
			name: "min length invalid",
			setup: func() nameregistry.Checker {
				return nameregistry.NewMinLengthChecker(4)
			},
			registerName: "x",
			errMsg:       "name is too short, min length is 4",
		},
		{
			name: "max length valid",
			setup: func() nameregistry.Checker {
				return nameregistry.NewMaxLengthChecker(4)
			},
			registerName: "x123",
		},
		{
			name: "max length invalid",
			setup: func() nameregistry.Checker {
				return nameregistry.NewMaxLengthChecker(4)
			},
			registerName: "xy123",
			errMsg:       "name is too long, max length is 4",
		},
		{
			name: "lowercase valid",
			setup: func() nameregistry.Checker {
				return nameregistry.NewLowercaseChecker()
			},
			registerName: "xyz123",
		},
		{
			name: "lowercase invalid",
			setup: func() nameregistry.Checker {
				return nameregistry.NewLowercaseChecker()
			},
			registerName: "XYZ123",
			errMsg:       "name characters must be in lowercase",
		},
		{
			name: "prefix valid with one prefix",
			setup: func() nameregistry.Checker {
				return nameregistry.NewForbiddenPrefixChecker("a")
			},
			registerName: "xyz",
		},
		{
			name: "prefix valid with many prefixes",
			setup: func() nameregistry.Checker {
				return nameregistry.NewForbiddenPrefixChecker("a", "b", "c")
			},
			registerName: "xyz",
		},
		{
			name: "prefix invalid with one prefix",
			setup: func() nameregistry.Checker {
				return nameregistry.NewForbiddenPrefixChecker("a")
			},
			registerName: "aaa",
			errMsg:       "name must not start these prefixes: a",
		},
		{
			name: "prefix invalid with many prefixes",
			setup: func() nameregistry.Checker {
				return nameregistry.NewForbiddenPrefixChecker("a", "b", "c")
			},
			registerName: "bbb",
			errMsg:       "name must not start these prefixes: a, b, c",
		},
		{
			name: "alphanum valid",
			setup: func() nameregistry.Checker {
				return nameregistry.NewAlphaNumChecker()
			},
			registerName: "x123",
		},
		{
			name: "alphanum valid with uppercase",
			setup: func() nameregistry.Checker {
				return nameregistry.NewAlphaNumChecker()
			},
			registerName: "X123",
		},
		{
			name: "alphanum invalid without start letter",
			setup: func() nameregistry.Checker {
				return nameregistry.NewAlphaNumChecker()
			},
			registerName: "123x",
			errMsg:       `name must start with a letter between "a" and "z"`,
		},
		{
			name: "alphanum invalid without numbers suffix",
			setup: func() nameregistry.Checker {
				return nameregistry.NewAlphaNumChecker()
			},
			registerName: "x",
			errMsg:       "name must end with three numbers",
		},
		{
			name: "alphanum invalid with few numbers suffix",
			setup: func() nameregistry.Checker {
				return nameregistry.NewAlphaNumChecker()
			},
			registerName: "x12",
			errMsg:       "name must end with three numbers",
		},
		{
			name: "homoglypgh valid",
			setup: func() nameregistry.Checker {
				return nameregistry.NewHomoglyphChecker(avl.NewTree())
			},
			registerName: "xyz",
		},
		{
			name: "homoglypgh of reserved name",
			setup: func() nameregistry.Checker {
				t := avl.NewTree()
				t.Set("foo", struct{}{})
				return nameregistry.NewHomoglyphChecker(t)
			},
			registerName: "foo",
			errMsg:       "name is reserved",
		},
		{
			name: "homoglypgh of similar reserved name",
			setup: func() nameregistry.Checker {
				t := avl.NewTree()
				t.Set("foo", struct{}{})
				return nameregistry.NewHomoglyphChecker(t)
			},
			registerName: "f00",
			errMsg:       "similar name already exists: foo",
		},
		{
			name: "homoglypgh of similar reserved name 2",
			setup: func() nameregistry.Checker {
				t := avl.NewTree()
				t.Set("foo", struct{}{})
				return nameregistry.NewHomoglyphChecker(t)
			},
			registerName: "foo000",
			errMsg:       "similar name already exists: foo",
		},
		{
			name: "homoglypgh of similar reserved name 3",
			setup: func() nameregistry.Checker {
				t := avl.NewTree()
				t.Set("foo", struct{}{})
				return nameregistry.NewHomoglyphChecker(t)
			},
			registerName: "foo000bar000",
			errMsg:       "similar name already exists: foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			checker := tt.setup()

			// Act
			err := checker.Check(tt.registerName)

			// Assert
			if tt.errMsg != "" {
				uassert.ErrorContains(t, err, tt.errMsg, "expect an error")
			} else {
				uassert.Nil(t, err, "expect no error")
			}
		})
	}
}
