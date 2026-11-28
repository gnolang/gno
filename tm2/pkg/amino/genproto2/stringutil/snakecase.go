// Copyright 2020-2023 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file has ToLowerSnakeCase which the Protobuf linter uses to make
// the expected spelling of a message field name. It is a copy of:
// https://github.com/bufbuild/buf/blob/main/private/pkg/stringutil/stringutil.go

package stringutil

import (
	"strings"
	"unicode"
)

// SnakeCaseOption is an option for snake_case conversions.
type SnakeCaseOption func(*snakeCaseOptions)

// SnakeCaseWithNewWordOnDigits is a SnakeCaseOption that signifies
// to split on digits, ie foo_bar_1 instead of foo_bar1.
func SnakeCaseWithNewWordOnDigits() SnakeCaseOption {
	return func(snakeCaseOptions *snakeCaseOptions) {
		snakeCaseOptions.newWordOnDigits = true
	}
}

// ToLowerSnakeCase transforms s to lower_snake_case.
func ToLowerSnakeCase(s string, options ...SnakeCaseOption) string {
	return strings.ToLower(toSnakeCase(s, options...))
}

func toSnakeCase(s string, options ...SnakeCaseOption) string {
	snakeCaseOptions := &snakeCaseOptions{}
	for _, option := range options {
		option(snakeCaseOptions)
	}
	output := ""
	s = strings.TrimFunc(s, isDelimiter)
	for i, c := range s {
		if isDelimiter(c) {
			c = '_'
		}
		if i == 0 {
			output += string(c)
		} else if isSnakeCaseNewWord(c, snakeCaseOptions.newWordOnDigits) &&
			output[len(output)-1] != '_' &&
			((i < len(s)-1 && !isSnakeCaseNewWord(rune(s[i+1]), true) && !isDelimiter(rune(s[i+1]))) ||
				(snakeCaseOptions.newWordOnDigits && unicode.IsDigit(c)) ||
				(unicode.IsLower(rune(s[i-1])))) {
			output += "_" + string(c)
		} else if !(isDelimiter(c) && output[len(output)-1] == '_') {
			output += string(c)
		}
	}
	return output
}

func isSnakeCaseNewWord(r rune, newWordOnDigits bool) bool {
	if newWordOnDigits {
		return unicode.IsUpper(r) || unicode.IsDigit(r)
	}
	return unicode.IsUpper(r)
}

func isDelimiter(r rune) bool {
	return r == '.' || r == '-' || r == '_' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

type snakeCaseOptions struct {
	newWordOnDigits bool
}
