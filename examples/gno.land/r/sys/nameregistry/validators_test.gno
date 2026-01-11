package nameregistry_test

import (
	"testing"

	"gno.land/p/nt/uassert"

	"gno.land/r/sys/nameregistry"
)

func TestNewValidator(t *testing.T) {
	uassert.PanicsWithMessage(t, "validator description is required", func() {
		_ = nameregistry.NewValidator("", nil)
	}, "expect required panic")

	uassert.PanicsWithMessage(t, "validator function is required", func() {
		_ = nameregistry.NewValidator("description", nil)
	}, "expect required panic")

	validator := nameregistry.NewValidator("Foobar", func(string) error { return nil })
	uassert.Equal(t, "Foobar", validator.Description(), "expect description to match")
}

func TestNewValidatorValidate(t *testing.T) {
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
			validator := nameregistry.NewValidator("Foobar", func(name string) error {
				validatedName = name
				return tt.err
			})

			// Act
			err := validator.Validate(tt.registerName)

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

func TestValidators(t *testing.T) {
	// Min length
	uassert.PanicsWithMessage(t, "min length validator requires a length > 0", func() {
		_ = nameregistry.NewMinLengthValidator(0)
	}, "expect invalid length panic")
	validator := nameregistry.NewMinLengthValidator(4)
	uassert.Equal(t, "Minimum name length allowed is 4", validator.Description(), "expect description to match")

	// Max length
	uassert.PanicsWithMessage(t, "max length validator requires a length > 1", func() {
		_ = nameregistry.NewMaxLengthValidator(0)
	}, "expect invalid length panic")
	validator = nameregistry.NewMaxLengthValidator(4)
	uassert.Equal(t, "Maximum name length allowed is 4", validator.Description(), "expect description to match")

	// Lowercase
	validator = nameregistry.NewLowercaseValidator()
	uassert.Equal(t, "Alphanumeric chars must all have lowercase chars", validator.Description(), "expect description to match")

	// Forbidden prefix
	uassert.PanicsWithMessage(t, "name prefix validator requires at least one prefix", func() {
		_ = nameregistry.NewForbiddenPrefixValidator()
	}, "expect panic when no prefix")
	validator = nameregistry.NewForbiddenPrefixValidator("a", "b")
	uassert.Equal(t, "Name must not start with: a, b", validator.Description(), "expect description to match")

	// Alphanum
	validator = nameregistry.NewAlphaNumValidator()
	uassert.Equal(t, "Name must start with a letter and end with 3 numbers", validator.Description(), "expect description to match")
}

func TestValidatorsValidate(t *testing.T) {
	tests := []struct {
		name         string
		registerName string
		setup        func() nameregistry.Validator
		errMsg       string
	}{
		{
			name: "min length valid",
			setup: func() nameregistry.Validator {
				return nameregistry.NewMinLengthValidator(4)
			},
			registerName: "x123",
		},
		{
			name: "min length invalid",
			setup: func() nameregistry.Validator {
				return nameregistry.NewMinLengthValidator(4)
			},
			registerName: "x",
			errMsg:       "name is too short, min length is 4",
		},
		{
			name: "max length valid",
			setup: func() nameregistry.Validator {
				return nameregistry.NewMaxLengthValidator(4)
			},
			registerName: "x123",
		},
		{
			name: "max length invalid",
			setup: func() nameregistry.Validator {
				return nameregistry.NewMaxLengthValidator(4)
			},
			registerName: "xy123",
			errMsg:       "name is too long, max length is 4",
		},
		{
			name: "lowercase valid",
			setup: func() nameregistry.Validator {
				return nameregistry.NewLowercaseValidator()
			},
			registerName: "xyz123",
		},
		{
			name: "lowercase invalid",
			setup: func() nameregistry.Validator {
				return nameregistry.NewLowercaseValidator()
			},
			registerName: "XYZ123",
			errMsg:       "name characters must be in lowercase",
		},
		{
			name: "prefix valid with one prefix",
			setup: func() nameregistry.Validator {
				return nameregistry.NewForbiddenPrefixValidator("a")
			},
			registerName: "xyz",
		},
		{
			name: "prefix valid with many prefixes",
			setup: func() nameregistry.Validator {
				return nameregistry.NewForbiddenPrefixValidator("a", "b", "c")
			},
			registerName: "xyz",
		},
		{
			name: "prefix invalid with one prefix",
			setup: func() nameregistry.Validator {
				return nameregistry.NewForbiddenPrefixValidator("a")
			},
			registerName: "aaa",
			errMsg:       "name must not start these prefixes: a",
		},
		{
			name: "prefix invalid with many prefixes",
			setup: func() nameregistry.Validator {
				return nameregistry.NewForbiddenPrefixValidator("a", "b", "c")
			},
			registerName: "bbb",
			errMsg:       "name must not start these prefixes: a, b, c",
		},
		{
			name: "alphanum valid",
			setup: func() nameregistry.Validator {
				return nameregistry.NewAlphaNumValidator()
			},
			registerName: "x123",
		},
		{
			name: "alphanum valid with uppercase",
			setup: func() nameregistry.Validator {
				return nameregistry.NewAlphaNumValidator()
			},
			registerName: "X123",
		},
		{
			name: "alphanum invalid without start letter",
			setup: func() nameregistry.Validator {
				return nameregistry.NewAlphaNumValidator()
			},
			registerName: "123x",
			errMsg:       `name must start with a letter between "a" and "z"`,
		},
		{
			name: "alphanum invalid without numbers suffix",
			setup: func() nameregistry.Validator {
				return nameregistry.NewAlphaNumValidator()
			},
			registerName: "x",
			errMsg:       "name must end with three numbers",
		},
		{
			name: "alphanum invalid with few numbers suffix",
			setup: func() nameregistry.Validator {
				return nameregistry.NewAlphaNumValidator()
			},
			registerName: "x12",
			errMsg:       "name must end with three numbers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			validator := tt.setup()

			// Act
			err := validator.Validate(tt.registerName)

			// Assert
			if tt.errMsg != "" {
				uassert.ErrorContains(t, err, tt.errMsg, "expect an error")
			} else {
				uassert.Nil(t, err, "expect no error")
			}
		})
	}
}
