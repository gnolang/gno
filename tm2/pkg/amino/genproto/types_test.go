package genproto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintP3Types(t *testing.T) {
	t.Parallel()

	doc := P3Doc{
		Comment: "doc comment",
		Messages: []P3Message{
			{
				Comment: "message comment",
				Name:    "message_name",
				Fields: []P3Field{
					{
						Comment:  "field_comment",
						Type:     P3ScalarTypeString,
						Name:     "field_name",
						Number:   1,
						Repeated: false,
					},
					{
						Comment:  "field_comment",
						Type:     P3ScalarTypeUint64,
						Name:     "field_name",
						Number:   2,
						Repeated: true,
					},
				},
			},
			{
				Comment: "message comment 2",
				Name:    "message_name_2",
				Fields:  []P3Field{},
			},
		},
	}

	proto3Schema := doc.Print()
	assert.Equal(t, `syntax = "proto3";

// doc comment

// messages
// message comment
message message_name {
	// field_comment
	string field_name = 1;
	// field_comment
	repeated uint64 field_name = 2;
}

// message comment 2
message message_name_2 {
}`, proto3Schema)
}
