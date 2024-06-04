package json

import (
	"errors"
	"io"
	"math"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
)

const (
	errUnexpectedEOF          = errors.New("unexpected EOF")
	errUnexpectedChar         = errors.New("unexpected character")
	errStringNotClosed        = errors.New("string not closed")
	errBracketNotClosed       = errors.New("bracket not closed")
	errInvalidSlicePathSyntax = errors.New("invalid slice path syntax")
	errInvalidSliceFromValue  = errors.New("invalid slice from value")
	errInvalidSliceToValue    = errors.New("invalid slice to value")
	errInvalidSliceStepValue  = errors.New("invalid slice step value")
)

// Path returns the nodes that match the given JSON path.
func Path(data []byte, path string) ([]*Node, error) {
	commands, err := parsePath(path)
	if err != nil {
		return nil, ufmt.Errorf("failed to parse path: %v", err)
	}

	nodes, err := Unmarshal(data)
	if err != nil {
		return nil, ufmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return applyPath(nodes, commands)
}

// parsePath parses the given path string and returns a slice of commands to be run.
//
// The function uses a state machine approach to parse the path based on the encountered tokens.
// It supports the following tokens and their corresponding states:
//   - Dollar sign ('$'): Appends a literal "$" to the result slice.
//   - Dot ('.'): Calls the processDot function to handle the dot token.
//   - Single dot ('.') followed by a child end character: Appends the substring between the dots to the result slice.
//   - Double dot ('..'): Appends ".." to the result slice.
//   - Opening bracket ('['): Calls the processBracketOpen function to handle the opening bracket token.
//   - Single quote ('”') after the opening bracket: Calls the processSingleQuote function to handle the string within single quotes.
//   - Any other character after the opening bracket: Calls the processWithoutSingleQuote function to handle the string without single quotes.
//
// The function returns the slice of parsed commands and any error encountered during the parsing process.
// If an unexpected character is encountered, an error (errUnexpectedChar) is returned.
func parsePath(path string) ([]string, error) {
	buf := newBuffer([]byte(path))
	result := make([]string, 0)

	for {
		b, err := buf.current()
		if err != nil {
			break
		}
		switch b {
		case dollarSign:
			result = append(result, "$")
		case atSign:
			result = append(result, "@")
		case dot:
			result, err = processDot(buf, result)
			if err != nil {
				return nil, err
			}
		case bracketOpen:
			result, err = processBracketOpen(buf, result)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errUnexpectedChar
		}

		err = buf.step()
		if err != nil && err != io.EOF {
			return nil, err
		}
	}

	return result, nil
}

func applyPath(node *Node, cmds []string) ([]*Node, error) {
	result := make([]*Node, 0)

	for i, cmd := range cmds {
		if i == 0 && (cmd == "$" || cmd == "@") { // root or current
			result = append(result, node)
			continue
		}

		var err error
		result, err = processCommand(cmd, result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// processCommand processes a single command on the given nodes.
//
// It determines the type of command and calls the corresponding function to handle the command.
func processCommand(cmd string, nodes []*Node) ([]*Node, error) {
	switch {
	case cmd == "..":
		return processRecursiveDescent(nodes), nil
	case cmd == "*":
		return processWildcard(nodes), nil
	case strings.Contains(cmd, ":"):
		return processSlice(cmd, nodes)
	case strings.HasPrefix(cmd, "?(") && strings.HasSuffix(cmd, ")"):
		panic("filter not implemented")
	default:
		res, err := processKeyUnion(cmd, nodes)
		if err != nil {
			return nil, err
		}

		return res, nil
	}
}

// processWildcard processes a wildcard command on the given nodes.
//
// It retrieves all the child nodes of each node in the given slice.
func processWildcard(nodes []*Node) (result []*Node) {
	for _, node := range nodes {
		result = append(result, node.getSortedChildren()...)
	}
	return result
}

// processRecursiveDescent performs a recursive descent on the given nodes.
//
// It recursively retrieves all the child nodes of each node in the given slice.
func processRecursiveDescent(nodes []*Node) (result []*Node) {
	for _, node := range nodes {
		result = append(result, recursiveChildren(node)...)
	}
	return result
}

// recursiveChildren returns all the recursive child nodes of the given node that are containers.
//
// It recursively traverses the child nodes of the given node and their child nodes,
// and returns a slice of pointers to all the child nodes that are containers.
func recursiveChildren(node *Node) (result []*Node) {
	if node.isContainer() {
		for _, element := range node.getSortedChildren() {
			if element.isContainer() {
				result = append(result, element)
			}
		}
	}

	temp := make([]*Node, 0, len(result))
	temp = append(temp, result...)

	for _, element := range result {
		temp = append(temp, recursiveChildren(element)...)
	}

	return temp
}

var pathSegmentDelimiters = map[byte]bool{dot: true, bracketOpen: true}

// processDot handles the processing when a dot character is found in the buffer.
//
// It checks the next character in the buffer.
// If the next character is also a dot, it appends ".." to the result slice.
// Otherwise, it reads the characters until the next child end character (childEnd) and appends the substring to the result slice.
// It returns the updated result slice and any error encountered.
func processDot(buf *buffer, result []string) ([]string, error) {
	start := buf.index

	b, err := buf.next()
	if err == io.EOF {
		err = nil
		return result, nil
	}

	if err != nil {
		return nil, err
	}

	if b == dot {
		result = append(result, "..")
		buf.index--
		return result, nil
	}

	err = buf.skipAny(pathSegmentDelimiters)
	stop := buf.index

	if err == io.EOF {
		err = nil
		stop = buf.length
	} else {
		buf.index--
	}

	if err != nil {
		return nil, err
	}

	if start+1 < stop {
		result = append(result, string(buf.data[start+1:stop]))
	}

	return result, nil
}

// processBracketOpen handles the processing when an opening bracket character ('[') is found in the buffer.
//
// It reads the next character in the buffer and determines the appropriate processing based on the character:
//   - If the next character is a single quote (`'`), it calls the processSingleQuote function to handle the string within single quotes.
//   - Otherwise, it calls the processWithoutSingleQuote function to handle the string without single quotes.
//
// It returns the updated result slice and any error encountered.
func processBracketOpen(buf *buffer, result []string) ([]string, error) {
	b, err := buf.next()
	if err != nil {
		return nil, errUnexpectedEOF
	}

	start := buf.index
	if b == singleQuote {
		result, err = processSingleQuote(buf, result, start)
	} else {
		result, err = processWithoutSingleQuote(buf, result, start)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// processSingleQuote handles the processing when a single quote character (`'`) is encountered after an opening bracket ('[') in the buffer.
//
// It assumes that the current position of the buffer is just after the single quote character.
//
// The function performs the following steps:
//  1. It skips the single quote character and reads the string until the next single quote character is found.
//  2. It checks if the character after the closing single quote is a closing bracket (']').
//     - If it is, the string between the single quotes is appended to the result slice.
//     - If it is not, an error (errBracketNotClosed) is returned.
//
// It returns the updated result slice and any error encountered.
func processSingleQuote(buf *buffer, result []string, start int) ([]string, error) {
	start++

	err := buf.string(singleQuote, true)
	if err != nil {
		return nil, errStringNotClosed
	}

	stop := buf.index

	b, err := buf.next()
	if err != nil {
		return nil, errUnexpectedEOF
	}

	if b != bracketClose {
		return nil, errBracketNotClosed
	}

	result = append(result, string(buf.data[start:stop]))

	return result, nil
}

// processWithoutSingleQuote handles the processing when a character other than
// a single quote (`'`) is encountered after an opening bracket ('[') in the buffer.
//
// It assumes that the current position of the buffer is just after the opening bracket.
//
// The function reads the characters until the next closing bracket (']') is found
// and appends the substring between the brackets to the result slice.
//
// It returns the updated result slice and any error encountered.
// If the closing bracket is not found, an error (errUnexpectedEOF) is returned.
func processWithoutSingleQuote(buf *buffer, result []string, start int) ([]string, error) {
	err := buf.skip(bracketClose)
	if err != nil {
		return nil, errUnexpectedEOF
	}

	stop := buf.index
	result = append(result, string(buf.data[start:stop]))

	return result, nil
}

// processSlice processes a slice path on the given nodes.
//
// The slice path has the following syntax:
//
//		[start:end:step]
//
//	  - start: The starting index of the slice (inclusive). if omitted, it defaults to 0.
//	    If negative, it counts from the end of the array.
//	  - end: The ending index of the slice (exclusive). if omitted, it defaults to the length of the array.
//	    If negative, it counts from the end of the array.
//	  - step: The step value for the slice. if omitted, it defaults to 1.
//
// The function performs the following steps:
//
// 1. Split the slice path into start, end, and step values.
//
// 2. Parses the each syntax components as integers.
//
// 3. For each node in the given nodes:
//   - If the node is an array:
//   - Calculate the length of the array.
//   - Adjust the start and end values if they are negative.
//   - Check if the slice range is within the bounds of the array.
//   - Iterate over the array elements based on the start, end and step values.
//   - Append the selected elements to the result slice.
//
// It returns the slice of selected nodes and any error encountered during the parsing process.
func processSlice(cmd string, nodes []*Node) ([]*Node, error) {
	from, to, step, err := parseSliceParams(cmd)
	if err != nil {
		return nil, err
	}

	var result []*Node
	for _, node := range nodes {
		if node.IsArray() {
			result = append(result, selectArrayElement(node, from, to, step)...)
		}
	}

	return result, nil
}

// parseSliceParams parses the slice parameters from the given path command.
func parseSliceParams(cmd string) (int64, int64, int64, error) {
	keys := strings.Split(cmd, ":")
	ks := len(keys)
	if ks > 3 {
		return 0, 0, 0, errInvalidSlicePathSyntax
	}

	from, err := strconv.ParseInt(keys[0], 10, 64)
	if err != nil {
		return 0, 0, 0, errInvalidSliceFromValue
	}

	to := int64(0)
	if ks > 1 {
		to, err = strconv.ParseInt(keys[1], 10, 64)
		if err != nil {
			return 0, 0, 0, errInvalidSliceToValue
		}
	}

	step := int64(1)
	if ks == 3 {
		step, err = strconv.ParseInt(keys[2], 10, 64)
		if err != nil {
			return 0, 0, 0, errInvalidSliceStepValue
		}
	}

	return from, to, step, nil
}

// selectArrayElement selects the array elements based on the given from, to, and step values.
func selectArrayElement(node *Node, from, to, step int64) []*Node {
	length := int64(len(node.next))

	if to == 0 {
		to = length
	}

	if from < 0 {
		from += length
	}

	if to < 0 {
		to += length
	}

	from = int64(math.Max(0, math.Min(float64(from), float64(length))))
	to = int64(math.Max(0, math.Min(float64(to), float64(length))))

	if step <= 0 || from >= to {
		return nil
	}

	// This formula calculates the number of elements that will be selected based on the given
	// from, to, and step values. It ensures that the correct number of elements are allocated
	// in the result slice.
	size := (to - from + step - 1) / step
	result := make([]*Node, 0, size)

	for i := from; i < to; i += step {
		if child, ok := node.next[ufmt.Sprintf("%d", i)]; ok {
			result = append(result, child)
		}
	}

	return result
}

// processKeyUnion processes a key union command on the given nodes.
//
// It retrieves the child nodes of each node in the given slice that match any of the specified keys
func processKeyUnion(cmd string, nodes []*Node) ([]*Node, error) {
	buf := newBuffer([]byte(cmd))
	keys, err := extractKeys(buf)
	if err != nil {
		return nil, err
	}

	var result []*Node
	for _, node := range nodes {
		if node.IsArray() {
			result, err = processArrayKeys(node, keys, result)
		} else if node.IsObject() {
			result, err = processObjectKeys(node, keys, result)
		}

		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func extractKeys(buf *buffer) ([]string, error) {
	keys := make([]string, 0)

	for {
		key, err := extractKey(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}

		keys = append(keys, key)

		if err := expectComma(buf); err != nil {
			return keys, nil
		}
	}

	return keys, nil
}

func extractKey(buf *buffer) (string, error) {
	from := buf.index
	err := buf.pathToken()
	if err != nil {
		return "", err
	}

	key := string(buf.data[from:buf.index])
	if len(key) > 2 && key[0] == singleQuote && key[len(key)-1] == singleQuote {
		key = key[1 : len(key)-1]
	}
	return key, nil
}

func expectComma(buf *buffer) error {
	c, err := buf.first()
	if err != nil {
		return err
	}

	if c != comma {
		return errUnexpectedChar
	}

	return buf.step()
}

func processArrayKeys(node *Node, keys []string, result []*Node) ([]*Node, error) {
	for _, key := range keys {
		index, err := strconv.Atoi(key)
		if err != nil {
			return nil, err
		}

		if index < 0 {
			index = node.Size() + index
		}

		if value, ok := node.next[ufmt.Sprintf("%d", index)]; ok {
			result = append(result, value)
		}
	}

	return result, nil
}

func processObjectKeys(node *Node, keys []string, result []*Node) ([]*Node, error) {
	for _, key := range keys {
		if value, ok := node.next[key]; ok {
			result = append(result, value)
		}
	}

	return result, nil
}

// Paths returns calculated paths of underlying nodes
func Paths(array []*Node) []string {
	result := make([]string, 0, len(array))
	for _, element := range array {
		result = append(result, element.Path())
	}

	return result
}
