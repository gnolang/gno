package ascii

import (
	"strings"
	"testing"
)

// padLine

func TestPadLine_Left(t *testing.T) {
	got := PadLine("hi", 6, AlignLeft, " ")
	want := "hi    "
	if got != want {
		t.Fatalf("padLine left failed: got '%s', want '%s'", got, want)
	}
}

func TestPadLine_Right(t *testing.T) {
	got := PadLine("hi", 6, AlignRight, " ")
	want := "    hi"
	if got != want {
		t.Fatalf("padLine right failed: got '%s', want '%s'", got, want)
	}
}

func TestPadLine_Center(t *testing.T) {
	got := PadLine("hi", 6, AlignWCenter, " ")
	want := "  hi  "
	if got != want {
		t.Fatalf("padLine center failed: got '%s', want '%s'", got, want)
	}
}

// padHeight

func TestPadHeight_Top(t *testing.T) {
	lines := []string{"one", "two"}
	out := PadHeight(lines, 4, AlignTop)
	if len(out) != 4 || out[0] != "one" || out[3] != "" {
		t.Fatalf("padHeight top failed: got %#v", out)
	}
}

func TestPadHeight_Center(t *testing.T) {
	lines := []string{"one"}
	out := PadHeight(lines, 3, AlignHCenter)
	if len(out) != 3 || out[0] != "" || out[1] != "one" || out[2] != "" {
		t.Fatalf("padHeight center failed: got %#v", out)
	}
}

func TestPadHeight_Bottom(t *testing.T) {
	lines := []string{"one"}
	out := PadHeight(lines, 2, AlignBottom)
	if len(out) != 2 || out[0] != "" || out[1] != "one" {
		t.Fatalf("padHeight bottom failed: got %#v", out)
	}
}

// Box

func TestBox_SingleLine(t *testing.T) {
	got := Box("Hi")
	if !strings.Contains(got, "| Hi |") {
		t.Fatalf("Box did not render expected line: got\n%s", got)
	}
}

func TestBox_MultiLineFallback(t *testing.T) {
	got := Box("One\nTwo")
	if !strings.Contains(got, "| One |") || !strings.Contains(got, "| Two |") {
		t.Fatalf("Box (multi-line fallback) failed: got\n%s", got)
	}
}

// FlexFrame

func TestFlexFrame(t *testing.T) {
	lines := []string{"Short", "A bit longer"}
	got := FlexFrame(lines, AlignRight)
	if !strings.Contains(got, "|        Short |") {
		t.Fatalf("FlexFrame right alignment failed:\n%s", got)
	}
}

// Frame

func TestFrame_SizeAlign(t *testing.T) {
	lines := []string{"line"}
	got := Frame(lines, AlignWCenter, 10, 3, AlignHCenter)
	if res := strings.Count(got, "\n"); res != 7 {
		t.Fatalf("Frame height failed (got %d):\n%s", res, got)
	}
	if !strings.Contains(got, "|    line    |") {
		t.Fatalf("Frame width or alignment failed:\n%s", got)
	}
}

// ProgressBar

func TestProgressBar(t *testing.T) {
	got := ProgressBar(30, 100, 10, true)
	if !strings.Contains(got, "30%") || !strings.Contains(got, "[###-------]") {
		t.Fatalf("ProgressBar output incorrect: %s", got)
	}
}

func TestProgressBar_ZeroTotal(t *testing.T) {
	got := ProgressBar(0, 0, 10, true)
	if !strings.Contains(got, "0%") {
		t.Fatalf("ProgressBar zero-total failed: %s", got)
	}
}

// Grid

func TestGrid(t *testing.T) {
	got := Grid(2, 3, "*")
	want := "***\n***\n"
	if got != want {
		t.Fatalf("Grid failed: got\n%s\nwant\n%s", got, want)
	}
}
