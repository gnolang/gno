package abi

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
)

func TestEncodeFunction(t *testing.T) {
	t.Parallel()

	num1 := big.NewInt(10)
	num2 := big.NewInt(20)
	expected := "ec2d2fd3000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000014"

	res, err := EncodeFunction("add(uint256,uint256)", num1, num2)
	if err != nil {
		t.Errorf("EncodeFunction returned an error: %v", err)
	}
	if fmt.Sprintf("%x", res) != expected {
		t.Errorf("EncodeFunction returned %x, expected %s", res, expected)
	}
}

func TestParseSignature(t *testing.T) {
	t.Parallel()

	funcName, paramTypes, err := parseSignature("add(uint256,uint256)")
	if err != nil {
		t.Errorf("parseSignature returned an error: %v", err)
	}
	if funcName != "add" {
		t.Errorf("parseSignature returned function name %s, expected 'add'", funcName)
	}
	if !reflect.DeepEqual(paramTypes, []string{"uint256", "uint256"}) {
		t.Errorf("parseSignature returned parameter types %v, expected ['uint256', 'uint256']", paramTypes)
	}
}

func TestEncodeParameter(t *testing.T) {
	t.Parallel()

	num := big.NewInt(10)
	expected := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa}

	res, err := encodeParameter("uint256", num)
	if err != nil {
		t.Errorf("encodeParameter returned an error: %v", err)
	}
	if !reflect.DeepEqual(res, expected) {
		t.Errorf("encodeParameter returned %x, expected %x", res, expected)
	}
}

func TestCalculateFunctionSelector(t *testing.T) {
	t.Parallel()

	expected := []byte{0x9a, 0x2c, 0xd6, 0x1f}

	res := calculateFunctionSelector("add", []string{"uint256", "uint256"})
	if !reflect.DeepEqual(res, expected) {
		t.Errorf("calculateFunctionSelector returned %x, expected %x", res, expected)
	}
}
