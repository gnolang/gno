package json

import (
	"testing"
)

func TestKeyExistsValidator(t *testing.T) {
	n, err := Unmarshal([]byte(`{"name": "John Doe", "age": 25, "city": "New York"}`))
	if err != nil {
		t.Errorf("KeyExistsValidator failed: %s", err)
	}

	validator := KeyExistsValidator{Key: "name"}
	valid, err := validator.Validate(n)
	if err != nil || !valid {
		t.Errorf("KeyExistsValidator failed: expected key 'name' to exist")
	}

	validator = KeyExistsValidator{Key: "age"}
	valid, err = validator.Validate(n)
	if err != nil || !valid {
		t.Errorf("KeyExistsValidator failed: expected key 'age' to exist")
	}

	validator = KeyExistsValidator{Key: "city"}
	valid, err = validator.Validate(n)
	if err != nil || !valid {
		t.Errorf("KeyExistsValidator failed: expected key 'city' to exist")
	}
}

func TestNumberRangeValidator(t *testing.T) {
	// TODO: NumberNode validator throws an stack overflow error when using Unmarshal
	// to create a *Node. Need to fix this.
	n := ObjectNode("", map[string]*Node{
		"age": NumberNode("age", 25),
	})

	validator := NumberRangeValidator{Key: "age", Min: 18, Max: 30}

	valid, err := validator.Validate(n)
}

func TestNumberRangeValidator_Fail(t *testing.T) {
	n := ObjectNode("", map[string]*Node{
		"age": NumberNode("age", 25),
	})

	validator := NumberRangeValidator{Key: "age", Min: 18, Max: 20}

	valid, err := validator.Validate(n)
	if valid {
		t.Errorf("TestNumberRangeValidator_Fail failed: expected age to be invalid. error: %s", err)
	}

	if err != nil {
		t.Errorf("NumberRangeValidator failed: %s", err)
	}
}

func TestValidateFunction(t *testing.T) {
	n, err := Unmarshal([]byte(`{"name": "John Doe", "age": 25, "city": "New York"}`))
	if err != nil {
		t.Errorf("Validate failed: %s", err)
	}

	vv := []Validator{
		KeyExistsValidator{Key: "name"},
		KeyExistsValidator{Key: "age"},
		KeyExistsValidator{Key: "city"},
		// NumberRangeValidator{Min: 18, Max: 30},
	}

	valid, err := Validate(n, vv...)
	if !valid {
		t.Errorf("Validate failed: expected age to be valid. error: %s", err)
	}
}
