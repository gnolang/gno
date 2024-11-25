package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

func main() {
	if len(os.Args) < 1 {
		fmt.Fprintln(os.Stderr, "invalid name")
		os.Exit(1)
	}

	name := os.Args[1]

	const width = 12
	if len(name) >= width {
		name = name[:width-3] + "..."
	}

	colorLeft := colorFromString(name, 0.5, 0.6, 90)
	colorRight := colorFromString(name, 1.0, 0.92, 90)
	borderStyle := lipgloss.NewStyle().Foreground(colorLeft).
		Border(lipgloss.ThickBorder(), false, true, false, false).
		BorderForeground(colorLeft).
		Bold(true).
		Width(width)
	lineStyle := lipgloss.NewStyle().Foreground(colorRight)

	w, r := os.Stdout, os.Stdin

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprint(w, borderStyle.Render(name)+" ")
		fmt.Fprintln(w, lineStyle.Render(line))
	}
}
