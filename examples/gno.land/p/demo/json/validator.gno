package json

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

// Validator is an interface that defines a method to validate a node.
type Validator interface {
	// Validate takes a node and returns whether the node is valid or not.
	Validate(*Node) (bool, error)
}

// KeyExistsValidator checks if a key exists in an object node.
type KeyExistsValidator struct {
	Key string
}

func (v KeyExistsValidator) Validate(n *Node) (bool, error) {
	if !n.IsObject() {
		return false, errors.New("node is not an object")
	}

	_, exists := n.next[v.Key]
	return exists, nil
}

// NumberRangeValidator checks if a node's value is within a specified range.
type NumberRangeValidator struct {
	Key string
	Min, Max float64
}

func (v NumberRangeValidator) Validate(n *Node) (bool, error) {
	if !n.IsObject() {
		return false, errors.New("expected an object node")
	}

	numberNode, exists := n.next[v.Key]
	if !exists {
		return false, ufmt.Errorf("key %s does not exist", v.Key)
	}

	if !numberNode.IsNumber() {
		return false, ufmt.Errorf("key %s is not a number", v.Key)
	}

	value, ok := numberNode.value.(float64)
	if !ok {
		return false, ufmt.Errorf("value of key %s is not a float64", v.Key)
	}

	return value >= v.Min && value <= v.Max, nil
}

// Validate takes a node and a list of validators and applies each validator to the node.
func Validate(m *Node, validators ...Validator) (bool, error) {
	for i, v := range validators {
		valid, err := v.Validate(m)
		if !valid {
			return false, err
		}
	}

	return true, nil
}