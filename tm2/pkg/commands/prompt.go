// Interactive prompt primitives for CLI wizard flows.
//
// This file provides reusable prompt functions that work with the commands.IO
// abstraction, making them testable with mock readers/writers and usable across
// any CLI tool in the Gno codebase.
//
// All prompt functions write to io.Err() (stderr) for prompts and error messages,
// and read from io.In() (stdin) for user input.
//
// Available prompts:
//   - [PromptString]: free-text input with optional default and validation
//   - [PromptChoice]: single-key choice menu (e.g. [r]ealm, [P]ackage)
//   - [PromptSelect]: numbered list menu with name matching
//
// Use [IsInteractive] to check whether stdin is a terminal before entering
// an interactive flow.

package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// IsInteractive returns true when standard input is a terminal,
// indicating that the program is running interactively and can
// prompt the user for input.
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Choice represents a single option in a PromptChoice menu.
type Choice struct {
	Aliases     []string // Additional accepted inputs (e.g. "realm", "package")
	Description string   // Human-readable label shown in the prompt
}

// PromptString prompts the user for a string value with an optional default.
// If validate is non-nil, the input is validated and the user is re-prompted
// on failure.
func PromptString(io IO, prompt string, defaultVal string, validate func(string) error) (string, error) {
	for {
		if defaultVal != "" {
			fmt.Fprintf(io.Err(), "%s [%s]: ", prompt, defaultVal)
		} else {
			fmt.Fprintf(io.Err(), "%s: ", prompt)
		}

		ans, err := readLine(io)
		if err != nil {
			return "", err
		}
		if ans == "" {
			ans = defaultVal
		}
		if validate != nil {
			if err := validate(ans); err != nil {
				fmt.Fprintf(io.Err(), "%s\n", err)
				continue
			}
		}
		return ans, nil
	}
}

// PromptChoice presents a single-key choice menu (e.g. "[r]ealm, [P]ackage, [m]ain")
// and returns the selected key. The map key is the short key the user types;
// defaultKey is the key selected on empty input (empty = no default).
func PromptChoice(io IO, prompt string, choices map[string]Choice, defaultKey string) (string, error) {
	for {
		fmt.Fprint(io.Err(), prompt)
		ans, err := readLine(io)
		if err != nil {
			return "", err
		}

		if ans == "" {
			if defaultKey != "" {
				return defaultKey, nil
			}
			fmt.Fprintf(io.Err(), "please enter a valid choice\n")
			continue
		}

		for key, c := range choices {
			if strings.EqualFold(key, ans) {
				return key, nil
			}
			for _, alias := range c.Aliases {
				if strings.EqualFold(alias, ans) {
					return key, nil
				}
			}
		}

		fmt.Fprintf(io.Err(), "invalid choice: %q\n", ans)
	}
}

// SelectItem represents a single option in a numbered PromptSelect menu.
type SelectItem struct {
	Name        string // Short name (e.g. "basic", "dao")
	Description string // One-line description shown next to the name
}

// PromptSelect presents a numbered list menu and returns the selected item's Name.
// Items are displayed in slice order (first = default, often most important).
// If there is only one item, it is auto-selected without prompting.
func PromptSelect(io IO, prompt string, items []SelectItem) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items available")
	}
	if len(items) == 1 {
		return items[0].Name, nil
	}

	for {
		fmt.Fprintf(io.Err(), "%s\n", prompt)
		for i, item := range items {
			fmt.Fprintf(io.Err(), "  %d. %s — %s\n", i+1, item.Name, item.Description)
		}
		fmt.Fprintf(io.Err(), "Choose [1]: ")

		ans, err := readLine(io)
		if err != nil {
			return "", err
		}
		if ans == "" {
			return items[0].Name, nil
		}

		if n, err := strconv.Atoi(ans); err == nil {
			if n >= 1 && n <= len(items) {
				return items[n-1].Name, nil
			}
			fmt.Fprintf(io.Err(), "invalid choice: %d (must be 1-%d)\n", n, len(items))
			continue
		}

		for _, item := range items {
			if strings.EqualFold(item.Name, ans) {
				return item.Name, nil
			}
		}
		fmt.Fprintf(io.Err(), "unknown choice: %q\n", ans)
	}
}

// readLine reads a single line from io.In() via GetString and trims whitespace.
func readLine(io IO) (string, error) {
	s, err := io.GetString("")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}
