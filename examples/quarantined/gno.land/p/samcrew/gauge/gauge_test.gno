package gauge

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name         string
		value        int
		total        int
		label        string
		color        string
		percentOnly  bool
		width        int
		canvasHeight int
		fontSize     int
		paddingH     int
		wantSubstr   string
	}{
		{
			name:         "fails when value <= 0",
			value:        0,
			total:        5,
			label:        "Yes",
			color:        "#a6da95",
			percentOnly:  false,
			width:        100,
			canvasHeight: 30,
			fontSize:     16,
			paddingH:     6,
			wantSubstr:   "jauge fails: value must be greater than 0",
		},
		{
			name:         "fails when total <= 0",
			value:        5,
			total:        0,
			label:        "Yes",
			color:        "#a6da95",
			percentOnly:  false,
			width:        100,
			canvasHeight: 30,
			fontSize:     16,
			paddingH:     6,
			wantSubstr:   "jauge fails: total must be greater than 0",
		},
		{
			name:         "fails when value > total",
			value:        11,
			total:        10,
			label:        "Yes",
			color:        "#a6da95",
			percentOnly:  false,
			width:        100,
			canvasHeight: 30,
			fontSize:     16,
			paddingH:     6,
			wantSubstr:   "jauge fails: value cannot be greater than total",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.value, tt.total, tt.label, tt.color, Config{tt.percentOnly, tt.width, tt.canvasHeight, tt.fontSize, tt.paddingH})
			if !strings.Contains(got, tt.wantSubstr) {
				t.Errorf("Render() = %q, want substring %q", got, tt.wantSubstr)
			}
		})
	}
}
