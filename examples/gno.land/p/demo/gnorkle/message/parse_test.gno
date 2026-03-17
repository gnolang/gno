package message_test

import (
	"testing"

	"gno.land/p/demo/gnorkle/message"
	"gno.land/p/nt/uassert/v0"
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

			uassert.Equal(t, string(tt.expFuncType), string(funcType))
			uassert.Equal(t, tt.expRemainder, remainder)
		})
	}
}
