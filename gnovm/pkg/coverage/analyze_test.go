package coverage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectExecutableLines(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
		want    map[int]bool
		wantErr bool
	}{
		{
			name: "Simple function",
			content: `
package main

func main() {
	x := 5
	if x > 3 {
		println("Greater")
	}
}`,
			want: map[int]bool{
				5: true, // x := 5
				6: true, // if x > 3
				7: true, // println("Greater")
			},
			wantErr: false,
		},
		{
			name: "Function with loop",
			content: `
package main

func loopFunction() {
	for i := 0; i < 5; i++ {
		if i%2 == 0 {
			continue
		}
		println(i)
	}
}`,
			want: map[int]bool{
				5: true, // for i := 0; i < 5; i++
				6: true, // if i%2 == 0
				7: true, // continue
				9: true, // println(i)
			},
			wantErr: false,
		},
		{
			name: "Only declarations",
			content: `
package main

import "fmt"

var x int

type MyStruct struct {
	field int
}`,
			want:    map[int]bool{},
			wantErr: false,
		},
		{
			name: "Invalid gno code",
			content: `
This is not valid Go code
It should result in an error`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := DetectExecutableLines(tt.content)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
