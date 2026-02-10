package forms

import (
	"testing"
)

func TestAnswerFormInvalidForm(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm("g1user"))
	db := NewDB()

	dataAllTypes := `[
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
	dataOneRequiredText := `[
		{
			"label": "Name",
			"fieldType": "string",
			"required": true
		}
	]`

	tests := []struct {
		name        string
		answer      string
		expectPanic bool
		data        string
	}{
		{
			name:        "correct",
			answer:      `["Alex", 21, true, 0, [0, 1]]`,
			expectPanic: false,
			data:        dataAllTypes,
		},
		{
			name:        "invalid string",
			answer:      `[0, 21, true, 0, [0, 1]`,
			expectPanic: true,
			data:        dataAllTypes,
		},
		{
			name:        "invalid number",
			answer:      `["Alex", "21", true, 0, [0, 1]]`,
			expectPanic: true,
			data:        dataAllTypes,
		},
		{
			name:        "invalid boolean",
			answer:      `["Alex", 21, 1, 0, [0, 1]]`,
			expectPanic: true,
			data:        dataAllTypes,
		},
		{
			name:        "invalid choice",
			answer:      `["Alex", 21, true, 10, [0, 1]]`,
			expectPanic: true,
			data:        dataAllTypes,
		},
		{
			name:        "invalid multi-choice 1",
			answer:      `["Alex", 21, true, 0, [0, 1, 2, 3, 4, 5]]`,
			expectPanic: true,
			data:        dataAllTypes,
		},
		{
			name:        "invalid multi-choice 2",
			answer:      `["Alex", 21, true, 0, [5]]`,
			expectPanic: true,
			data:        dataAllTypes,
		},
		{
			name:        "invalid multi-choice 3",
			answer:      `["Alex", 21, true, 0, 0]`,
			expectPanic: true,
			data:        dataAllTypes,
		},
		{
			name:        "required string",
			answer:      `["", 21, true, 0, [0, 1]]`,
			expectPanic: true,
			data:        dataAllTypes,
		},
		{
			name:        "unrequired number",
			answer:      `["Alex", null, true, 0, [0, 1]]`,
			expectPanic: false,
			data:        dataAllTypes,
		},
		{
			name:        "correct one field",
			answer:      `["Alex"]`,
			expectPanic: false,
			data:        dataOneRequiredText,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formID, err := db.CreateForm("Test Form", "Test Description", "", "", tt.data)
			if err != nil {
				t.Error(err)
			}

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
