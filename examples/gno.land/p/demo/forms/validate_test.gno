package forms

import (
	"testing"
)

func TestAnswerFormInvalidForm(t *testing.T) {
	db := NewDatabase()

	data := `[
		{
			"label": "Name",
			"fieldType": "string",
			"required": true
		},
		{
			"label": "Age",
			"fieldType": "number",
			"required": false
		},
		{
			"label": "Is this a test?",
			"fieldType": "boolean",
			"required": false
		},
		{
			"label": "Favorite Food",
			"fieldType": "[Pizza|Schnitzel|Burger]",
			"required": true
		},
		{
			"label": "Favorite Foods",
			"fieldType": "{Pizza|Schnitzel|Burger}",
			"required": true
		}
	]`

	formID := db.CreateForm("Test Form", "Test Description", "2020-01-01", "2020-01-02", data)

	tests := []struct {
		name        string
		answer      string
		expectPanic bool
	}{
		{
			name:        "correct",
			answer:      `["Alex", 21, true, 0, [0, 1]]`,
			expectPanic: false,
		},
		{
			name:        "invalid string",
			answer:      `[0, 21, true, 0, [0, 1]`,
			expectPanic: true,
		},
		{
			name:        "invalid number",
			answer:      `["Alex", "21", true, 0, [0, 1]]`,
			expectPanic: true,
		},
		{
			name:        "invalid boolean",
			answer:      `["Alex", 21, 1, 0, [0, 1]]`,
			expectPanic: true,
		},
		{
			name:        "invalid choice",
			answer:      `["Alex", 21, true, 10, [0, 1]]`,
			expectPanic: true,
		},
		{
			name:        "invalid multi-choice 1",
			answer:      `["Alex", 21, true, 0, [0, 1, 2, 3, 4, 5]]`,
			expectPanic: true,
		},
		{
			name:        "invalid multi-choice 2",
			answer:      `["Alex", 21, true, 0, [5]]`,
			expectPanic: true,
		},
		{
			name:        "invalid multi-choice 3",
			answer:      `["Alex", 21, true, 0, 0]`,
			expectPanic: true,
		},
		{
			name:        "required string",
			answer:      `["", true, 0, [0, 1]]`,
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if tt.expectPanic {
						return
					}
					t.Errorf("%s panic occurred when not expected: %v", tt.name, r)
				} else if tt.expectPanic {
					t.Errorf("%s expected panic but didn't occur", tt.name)
				}
			}()

			db.SubmitForm(formID, tt.answer)
		})
	}
}
