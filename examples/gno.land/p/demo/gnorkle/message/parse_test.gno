package message_test

import (
	"testing"

	"gno.land/p/demo/gnorkle/message"
)

func TestParseFunc(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expFuncType  message.FuncType
		expRemainder string
	}{
		{
			name: "empty",
		},
		{
			name:        "func only",
			input:       "ingest",
			expFuncType: message.FuncTypeIngest,
		},
		{
			name:         "func with short remainder",
			input:        "commit,asdf",
			expFuncType:  message.FuncTypeCommit,
			expRemainder: "asdf",
		},
		{
			name:         "func with long remainder",
			input:        "request,hello,world,goodbye",
			expFuncType:  message.FuncTypeRequest,
			expRemainder: "hello,world,goodbye",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcType, remainder := message.ParseFunc(tt.input)
			if funcType != tt.expFuncType {
				t.Errorf("expected func type %s, got %s", tt.expFuncType, funcType)
			}

			if remainder != tt.expRemainder {
				t.Errorf("expected remainder of %s, got %s", tt.expRemainder, remainder)
			}
		})
	}
}
