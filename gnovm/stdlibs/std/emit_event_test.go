package std

import (
	"fmt"
	"testing"
	"time"
)

func TestNewEventWithContext(t *testing.T) {
	t.Parallel()

	currentTimestamp := time.Now().Unix()
	pkgPath := "gno.land/p/demo"

	tests := []struct {
		name       string
		ctx        *ExecContext
		eventType  string
		pkgPath    string
		attributes []string
		wantErr    bool
	}{
		{
			name: "valid event with attributes",
			ctx: &ExecContext{
				Height:    10,
				Timestamp: currentTimestamp,
			},
			eventType:  "foo",
			pkgPath:    pkgPath,
			attributes: []string{"key1", "value1", "key2", "value2"},
			wantErr:    false,
		},
		{
			name: "invalid event with odd number of attributes",
			ctx: &ExecContext{
				Height:    20,
				Timestamp: currentTimestamp,
			},
			eventType:  "bar",
			pkgPath:    pkgPath,
			attributes: []string{"key1", "value1", "key2"},
			wantErr:    true,
		},
		{
			name: "empty event type",
			ctx: &ExecContext{
				Height:    30,
				Timestamp: currentTimestamp,
			},
			eventType:  "",
			pkgPath:    pkgPath,
			attributes: []string{"key1", "value1"},
			wantErr:    true,
		},
		{
			name: "long attributes list",
			ctx: &ExecContext{
				Height:    40,
				Timestamp: currentTimestamp,
			},
			eventType:  "longAttr",
			pkgPath:    pkgPath,
			attributes: makeLongAttributesList(50),
			wantErr:    false,
		},
		{
			name: "duplicate keys",
			ctx: &ExecContext{
				Height:    50,
				Timestamp: currentTimestamp,
			},
			eventType:  "dupKeys",
			pkgPath:    pkgPath,
			attributes: []string{"key1", "value1", "key1", "value2"},
			wantErr:    false,
		},
		{
			name: "special characters in keys and values",
			ctx: &ExecContext{
				Height:    60,
				Timestamp: currentTimestamp,
			},
			eventType:  "specialChars",
			pkgPath:    pkgPath,
			attributes: []string{"key1$", "value@1", "key^2", "value#2"},
			wantErr:    false,
		},
		{
			name: "very large height and timestamp",
			ctx: &ExecContext{
				Height:    9223372036854775807, // Max int64 value
				Timestamp: 9223372036854775807,
			},
			eventType:  "largeValues",
			pkgPath:    pkgPath,
			attributes: []string{"key1", "value1"},
			wantErr:    false,
		},
		{
			name: "attributes without event",
			ctx: &ExecContext{
				Height:    70,
				Timestamp: currentTimestamp,
			},
			eventType:  "noAttrs",
			pkgPath:    pkgPath,
			attributes: nil,
			wantErr:    false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			event, err := NewEvent(tc.ctx, tc.eventType, tc.pkgPath, tc.attributes...)
			if (err != nil) != tc.wantErr {
				t.Fatalf("NewEventWithCtx() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err == nil {
				if event.Type != tc.eventType {
					t.Errorf("expected eventType %s, got %s", tc.eventType, event.Type)
				}
				if event.PkgPath != tc.pkgPath {
					t.Errorf("expected pkgPath %s, got %s", tc.pkgPath, event.PkgPath)
				}
				if event.Height != tc.ctx.Height {
					t.Errorf("expected height %d, got %d", tc.ctx.Height, event.Height)
				}
				if event.Timestamp != tc.ctx.Timestamp {
					t.Errorf("expected timestamp %d, got %d", tc.ctx.Timestamp, event.Timestamp)
				}
			}
		})
	}
}

func makeLongAttributesList(n int) []string {
	attrs := make([]string, 0, n*2)
	for i := 0; i < n; i++ {
		attrs = append(attrs, "key"+fmt.Sprintf("%d", i), "value"+fmt.Sprintf("%d", i))
	}
	return attrs
}

func TestAddAttributes(t *testing.T) {
	t.Parallel()
	ctx := &ExecContext{
		Height:    10,
		Timestamp: time.Now().Unix(),
	}

	event, err := NewEvent(ctx, "foo", "gno.land/p/demo", "key1", "value1")
	if err != nil {
		t.Fatalf("NewEvent() error = %v", err)
	}

	event.AddAttribute("key2", "value2")

	if len(event.Attributes) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(event.Attributes))
	}

	if event.Attributes[0].Key != "key1" || event.Attributes[0].Value != "value1" {
		t.Errorf("expected key1:value1, got %s:%s", event.Attributes[0].Key, event.Attributes[0].Value)
	}

	if event.Attributes[1].Key != "key2" || event.Attributes[1].Value != "value2" {
		t.Errorf("expected key2:value2, got %s:%s", event.Attributes[1].Key, event.Attributes[1].Value)
	}
}
