package abi

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"
)

// EncodeFunction encodes a function call with the given signature and parameters into an ABI encoding.
func EncodeFunction(signature string, params ...interface{}) ([]byte, error) {
	funcName, paramTypes, err := parseSignature(signature)
	if err != nil {
		return nil, err
	}

	if len(paramTypes) != len(params) {
		return nil, fmt.Errorf("number of parameters does not match the signature")
	}

	encodedParams := make([]byte, 0)
	for i, paramType := range paramTypes {
		param := params[i]
		encodedParam, err := encodeParameter(paramType, param)
		if err != nil {
			return nil, err
		}
		encodedParams = append(encodedParams, encodedParam...)
	}

	selector := calculateFunctionSelector(funcName, paramTypes)
	abiEncoding := append(selector, encodedParams...)

	return abiEncoding, nil
}

// parseSignature parses a function signature and returns the function name and parameter types.
func parseSignature(signature string) (string, []string, error) {
	// validate function signature format
	if !strings.Contains(signature, "(") || !strings.Contains(signature, ")") {
		return "", nil, fmt.Errorf("invalid function signature")
	}

	// split the function name and param parts
	parts := strings.Split(signature, "(")
	funcName := parts[0]
	paramsString := strings.TrimSuffix(parts[1], ")")

	// extract param type
	paramTypes := make([]string, 0)
	if paramsString != "" {
		paramTypes = strings.Split(paramsString, ",")
		for i := range paramTypes {
			paramTypes[i] = strings.TrimSpace(paramTypes[i])
		}
	}

	return funcName, paramTypes, nil
}

// encodeParameter encodes a parameter value based on its type.
func encodeParameter(paramType string, param interface{}) ([]byte, error) {
	switch paramType {
	case "uint256":
		num, ok := param.(*big.Int) // TODO: remove bigInt
		if !ok {
			return nil, fmt.Errorf("invalid parameter type for uint256")
		}
		return encodeUint256(num), nil

	case "address": // std.Address
		addr, ok := param.([]byte)
		if !ok || len(addr) != 20 {
			return nil, fmt.Errorf("invalid parameter type for address")
		}
		return encodeAddress(addr), nil

	case "bool":
		b, ok := param.(bool)
		if !ok {
			return nil, fmt.Errorf("invalid parameter type for bool")
		}
		return encodeBool(b), nil

	case "string":
		str, ok := param.(string)
		if !ok {
			return nil, fmt.Errorf("invalid parameter type for string")
		}
		return encodeString(str), nil

	// TODO: add more types

	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", paramType)
	}
}

// encodeUint256 encodes a uint256 value into a 32-byte big-endian byte slice.
func encodeUint256(num *big.Int) []byte {
	bytes := make([]byte, 32)
	num.FillBytes(bytes)

	return bytes
}

// encodeAddress encodes an address value into a 32-byte byte slice.
func encodeAddress(addr []byte) []byte {
	bytes := make([]byte, 32)
	copy(bytes[12:], addr)

	return bytes
}

// encodeBool encodes a boolean value into a 32-byte byte slice.
func encodeBool(b bool) []byte {
	bytes := make([]byte, 32)
	if b {
		bytes[31] = 1
	}

	return bytes
}

// encodeString encodes a string value into a byte slice with a 32-byte length prefix.
func encodeString(str string) []byte {
	strBytes := []byte(str)
	lenBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(lenBytes[24:], uint64(len(strBytes)))
	return append(lenBytes, strBytes...)
}

// calculateFunctionSelector calculates the function selector based on the function name and parameter types.
func calculateFunctionSelector(funcName string, paramTypes []string) []byte {
	var builder strings.Builder

	builder.WriteString(funcName)
	builder.WriteString("(")

	for i, paramType := range paramTypes {
		builder.WriteString(paramType)
		if i < len(paramTypes)-1 {
			builder.WriteString(",")
		}
	}

	builder.WriteString(")")
	signature := builder.String()

	hash := sha256.Sum256([]byte(signature))
	selector := hash[:4]

	return selector
}
