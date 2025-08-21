package integration

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// `unquote` takes a slice of strings, resulting from splitting a string block by spaces, and
// processes them. The function handles quoted phrases and escape characters within these strings.
func unquote(args []string) ([]string, error) {
	const quote = '"'

	parts := []string{}
	var inQuote bool

	var part strings.Builder
	for _, arg := range args {
		var escaped bool
		for _, c := range arg {
			if escaped {
				// If the character is meant to be escaped, it is processed with Unquote.
				// We use `Unquote` here for two main reasons:
				// 1. It will validate that the escape sequence is correct
				// 2. It converts the escaped string to its corresponding raw character.
				//    For example, "\\t" becomes '\t'.
				uc, err := strconv.Unquote(`"\` + string(c) + `"`)
				if err != nil {
					return nil, fmt.Errorf("unhandled escape sequence `\\%c`: %w", c, err)
				}

				part.WriteString(uc)
				escaped = false
				continue
			}

			// If we are inside a quoted string and encounter an escape character,
			// flag the next character as `escaped`
			if inQuote && c == '\\' {
				escaped = true
				continue
			}

			// Detect quote and toggle inQuote state
			if c == quote {
				inQuote = !inQuote
				continue
			}

			// Handle regular character
			part.WriteRune(c)
		}

		// If we're inside a quote, add a single space.
		// It reflects one or multiple spaces between args in the original string.
		if inQuote {
			part.WriteRune(' ')
			continue
		}

		// Finalize part, add to parts, and reset for next part
		parts = append(parts, part.String())
		part.Reset()
	}

	// Check if a quote is left open
	if inQuote {
		return nil, errors.New("unfinished quote")
	}

	return parts, nil
}

// splitArgs splits a flags line into arguments, respecting both single and
// double quotes and matching shell-like conventions.
// Returns an error if there is an unfinished quote.
func splitArgs(s string) ([]string, error) {
	const (
		singleQuote = '\''
		doubleQuote = '"'
	)

	var (
		cur       strings.Builder
		inQuote   bool
		quoteChar rune // Either ' or "
		escape    bool
	)

	args := []string{}
	for _, r := range s {
		switch {
		case escape:
			// Always treat next character as literal when escaping (except in single quotes)
			cur.WriteRune(r)
			escape = false

		case r == '\\':
			// Only enable escaping outside single quotes
			if inQuote && quoteChar == singleQuote {
				cur.WriteRune(r)
			} else {
				escape = true
			}

		case r == singleQuote || r == doubleQuote:
			if !inQuote {
				inQuote = true
				quoteChar = r
			} else if quoteChar == r {
				inQuote = false
				quoteChar = 0
			} else {
				// Different quote inside quoted string, treat as literal
				cur.WriteRune(r)
			}

		case r == ' ' && !inQuote:
			// End of an argument
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}

		default:
			cur.WriteRune(r)
		}
	}

	if inQuote {
		return nil, errors.New("unfinished quote")
	}
	if escape {
		// Trailing backslash at end of input
		cur.WriteRune('\\')
	}

	if cur.Len() > 0 {
		args = append(args, cur.String())
	}

	return args, nil
}
