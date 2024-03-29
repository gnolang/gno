package std

import (
	"strconv"
	"testing"
)

type testEvent struct {
	name      string
	eventType string
	attrs     []string
	expected  map[string]string
	expectErr bool
}

func TestCreateNewEventAndAddAttributesWithStringSlice(t *testing.T) {
	t.Parallel()

	cases := []testEvent{
		{
			name:      "Basic event with attributes",
			eventType: "hello",
			attrs:     []string{"world", "hello world!", "foo", "bar"},
			expected: map[string]string{
				"Type":       "hello",
				"Attributes": "2",
				"world":      "hello world!",
				"foo":        "bar",
			},
		},
		{
			name:      "Event with odd number of attributes",
			eventType: "hello",
			attrs:     []string{"world", "hello world!", "foo"},
			expectErr: true,
		},
		{
			name:      "Event with no attributes",
			eventType: "hello",
			expected: map[string]string{
				"Type":       "hello",
				"Attributes": "0",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, err := NewEvent(tc.eventType, tc.attrs...)
			actual := map[string]string{
				"Type":       e.Type,
				"Attributes": strconv.Itoa(len(e.Attributes)),
			}

			for _, attr := range e.Attributes {
				actual[attr.Key] = attr.Value
			}

			for key, expectedValue := range tc.expected {
				if actual[key] != expectedValue {
					t.Errorf("expected %s for key %s, got %s", expectedValue, key, actual[key])
				}
			}

			if tc.expectErr && err == nil {
				t.Fatalf("expected error, got %+v", err)
			}
		})
	}
}

func TestCreateNewEventWithUsingAddAttribute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		attributes map[string]string
		expected   map[string]string
	}{
		{
			name: "Test 1",
			attributes: map[string]string{
				"world": "hello world!",
				"foo":   "bar",
			},
			expected: map[string]string{
				"Type":       "hello",
				"Attributes": "2",
				"world":      "hello world!",
				"foo":        "bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := NewEvent("hello")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			for key, value := range tt.attributes {
				e.AddAttribute(key, value)
			}

			actual := map[string]string{
				"Type":       e.Type,
				"Attributes": strconv.Itoa(len(e.Attributes)),
				"world":      e.Attributes[0].Value,
				"foo":        e.Attributes[1].Value,
			}

			for key, value := range tt.expected {
				if actual[key] != value {
					t.Errorf("expected %s, got %s", value, actual[key])
				}
			}
		})
	}
}
